<script lang="ts">
  import Client from "./components/Client.svelte";
  import Login from "./components/Login.svelte";
  import Authorize from "./components/Authorize.svelte";
  import Connect from "./components/Connect.svelte";
  import AddServerDialog from "./components/AddServerDialog.svelte";
  import ServerRail, { type RailServer } from "./components/ServerRail.svelte";
  import type { RemoteVoice } from "./components/Client.svelte";
  import { inVoice, isLobby, joinChannel, setSelfState } from "@kallisti/protocol";
  import { createAccountsApi, type Account, type AccountActions } from "./lib/accounts";
  import { connect, type Connection, type Credentials } from "./lib/connection.svelte";
  import { createServerSession, type ServerSession, type Target } from "./lib/serverSession.svelte";
  import { createServerHub, type HubServer } from "./lib/serverHub.svelte";
  import { iconHref } from "./lib/serverLogo";
  import {
    bridgeUrl,
    getBridge,
    isDesktop,
    isKallisti,
    kallistiOrigin,
    kallistiSocket,
    linuxApp,
    parseAddress,
    trustServer,
  } from "./lib/desktop";
  import { httpBase } from "./lib/textChannels";
  import { serverOrigin } from "./lib/serverList";
  import { safeGet, safeSet, tabStorage } from "./lib/storage";
  import { proxyUrl } from "./lib/status";
  import { loggedOut } from "./lib/reconnect";
  import { untrack, type Snippet } from "svelte";
  import { checkForUpdate } from "./lib/update";
  import { browserPushSupport, dropServerPushes, pushIsOn, syncServerPush, turnOffPush, type PushPrefs } from "./lib/push";
  import { voiceJoined } from "./lib/voice";
  import { inviteOf, inviteRoute } from "./lib/invites";
  import type { RailBadge } from "./lib/title";
  import { voiceSupported } from "./lib/voice/support";
  import { loadCaptureWorklet, startAfter } from "./lib/voice/worklet";
  import { closeVoiceOutput } from "./lib/voice/output";
  import { swapContext } from "./lib/voice/restart";
  import { settings } from "./lib/settings.svelte";
  import { themeOf } from "./lib/settings";
  import { fediResult, reauthInstance, withoutFediResult } from "./lib/fedi";
  import Toast from "./components/Toast.svelte";

  // The colour theme (see app.css), also before the login.
  const systemDark = matchMedia("(prefers-color-scheme: dark)");
  let isSystemDark = $state(systemDark.matches);
  $effect(() => {
    const update = () => (isSystemDark = systemDark.matches);
    systemDark.addEventListener("change", update);
    return () => systemDark.removeEventListener("change", update);
  });
  $effect(() => {
    const theme = themeOf(settings.value.theme, isSystemDark);
    document.documentElement.dataset.theme = theme;
    const rail = getComputedStyle(document.documentElement).getPropertyValue("--bg-rail").trim();
    document.querySelector('meta[name="theme-color"]')?.setAttribute("content", rail);
  });

  const PROXY_URL = import.meta.env.VITE_PROXY_URL ?? proxyUrl(location);

  // The account API lives on the Kallisti server (same host as the bridge).
  // The desktop app's server is chosen on the login screen (see below).
  let api = $state.raw(createAccountsApi(httpBase(PROXY_URL)));
  let account = $state<Account | null>(null);
  // Until the server says whether a session is saved: no login screen (it
  // would flash). A saved session connects at once; logging out is the way
  // to another account. Only if the user has not acted yet: a late answer
  // must not undo a login. The desktop app asks its server when one is
  // chosen (see chooseServer).
  let checking = $state(true);
  // The home login is known for sure (a check answered, or the user logged
  // in or out): only then may another owner end the added servers (see
  // lib/serverHub follow). A failed check knows nothing.
  let ownerKnown = $state(false);
  // Each login, logout and server change: older account checks count for
  // nothing (see the check after an error).
  let authGeneration = 0;
  const autoLogin = (acc: Account | null) => {
    // The popup of a handoff logs in nobody here.
    if (!acc || connection || account !== null || connectHome) return void (checking = false);
    account = acc;
    // Another server asks who you are: only the question, no chat (a Mumble
    // connection would take over this user's voice elsewhere).
    if (authorizeQuery) return void (checking = false);
    // Still "checking" until the connection exists (start is async).
    void start({ username: acc.name, account: true, ...(desktop && { server: desktopServer }) }).finally(() => (checking = false));
  };
  // An account connection failed (for example an expired session): ask the
  // server again, so the login form comes back instead of a dead end.
  $effect(() => {
    // Unknown (the server or the network failed): no change; an answer
    // from before a login, logout or server change counts for nothing.
    if (session.error && connection?.account) {
      const mine = authGeneration;
      void api.check().then((acc) => {
        if (acc !== "unknown" && mine === authGeneration) account = acc;
      });
    }
  });


  // Shown on the login screen after leaving (for example a failed logout).
  let loginNote = $state<string | null>(null);
  // Back from a fediverse login (see lib/fedi): its result, once (a reload
  // must not show it again).
  const fediBack = isDesktop() ? null : fediResult(location.search);
  // Another server asks who you are (see Authorize): after the login here.
  const authorizeQuery = isDesktop() ? null : new URLSearchParams(location.search).get("authorize");
  // This page is the popup in which a Kallisti page adds this server (see
  // Connect): its home server, kept in this tab over the login's pages.
  // Such a page is only that popup: no account here, no servers of its own.
  const popupLogin = (() => {
    const tab = tabStorage();
    if (isDesktop()) return { home: null, kept: true, resumed: false, tab };
    const asked = new URLSearchParams(location.search).get("kallisti-connect");
    // kept: it survives the login's pages (else it could never finish).
    // It stays until the token is delivered, fails or is given up (the
    // tab goes with the popup).
    const kept = !asked || safeSet(tab, "kallisti-connect", asked);
    const mark = safeGet(tab, "kallisti-connect") || null;
    // A reload in the name step: no result in the address any more.
    const resumed = !asked && !fediBack && !!mark;
    const back = (fediBack && ["connect", "register", "error"].includes(fediBack.kind)) || resumed ? mark : null;
    return { home: asked ?? back, kept, resumed, tab };
  })();
  const connectHome = popupLogin.home;
  let connectPhase = $state<"start" | "deliver" | { error: string } | null>(
    !connectHome
      ? null
      : !popupLogin.kept
        ? { error: "This browser keeps no data for this window. Log in with a password there instead." }
        : popupLogin.resumed
          ? null // the name step again
          : !fediBack
            ? "start"
            : fediBack.kind === "connect"
              ? "deliver"
              : fediBack.kind === "error"
                ? { error: fediBack.text }
                : null, // register: the name first (Login), then deliver
  );
  // Delivered or failed: a later login in this tab is a normal one.
  $effect(() => {
    if (connectPhase !== null && connectPhase !== "start") safeSet(popupLogin.tab, "kallisti-connect", "");
  });
  // Back from a login through another server with an invite (see Login):
  // its address again, so the client redeems it.
  // Kept just before the page went to that login (a known account comes
  // back without a result in the address): used by the next page, once.
  const keptInvite = safeGet(popupLogin.tab, "kallisti-invite");
  if (keptInvite) safeSet(popupLogin.tab, "kallisti-invite", "");
  const restore = keptInvite?.startsWith("/invite/") || keptInvite?.startsWith("/?join=") ? keptInvite : null;
  if (restore) history.replaceState(history.state, "", restore);
  else if (fediBack) history.replaceState(history.state, "", withoutFediResult(location.pathname + location.search + location.hash));
  let fediSignup = $state(fediBack?.kind === "register" || popupLogin.resumed);
  let fediNotice = $state<string | null>(null);
  // Back from a fresh login for the first password: to the settings again.
  let reopenSettings = $state(fediBack?.kind === "reauth");
  if (!isDesktop())
    void untrack(() => api)
      .check()
      .then((answer) => {
        // The popup of a handoff: nobody's login here counts (it must not
        // end this browser's other tabs' servers, nor the name step).
        if (connectHome) return void (checking = false);
        // Unknown (the server or the network failed): the login form, but
        // nothing of the last login ends (see ownerKnown).
        if (answer !== "unknown") ownerKnown = true;
        const acc = answer === "unknown" ? null : answer;
        // Logged in: a notice over the app (and no name step); else errors
        // on the login screen.
        if (acc) fediSignup = false;
        if (fediBack && "text" in fediBack) {
          if (acc) fediNotice = fediBack.text;
          else if (fediBack.kind === "error") loginNote = fediBack.text;
        }
        autoLogin(acc);
      });
  // endAdded: the home login ends, and the added servers with it (see
  // lib/serverHub); not for a refused certificate (no login ended).
  const leave = (note: string | null, endAdded = true) => {
    authGeneration++;
    loginNote = note;
    addingFor = null;
    if (endAdded) hub.endAll();
    session.leave();
    closeAudio(audio);
    audio = null;
  };
  // The session ended on the server (a logout or password change on
  // another device): back to the login screen, as after a logout. The
  // added servers only close: whoever ended it ended their tokens (a tab
  // that wakes late must not end a newer login's).
  $effect(() => {
    if (connection?.closed && loggedOut(connection.closeReason))
      untrack(() => {
        hub.detach();
        account = null;
        void stopPush();
        leave("You were logged out.", false);
      });
  });
  // This browser's push notifications end with the login (see lib/push).
  const stopPush = () => (browserPushSupport() === "ok" ? turnOffPush(httpBase(PROXY_URL)) : Promise.resolve());
  const accountActions: AccountActions = {
    get temporary() {
      return account?.temporary ?? false;
    },
    // An answer that comes after another login or logout acts on nothing.
    logout: async () => {
      void stopPush(); // not waited for: the logout ends it on the server too
      const mine = authGeneration;
      const ok = await api.logout({ session: account?.session, name: account?.name });
      if (mine !== authGeneration) return;
      if (ok === "other") {
        // Another login is here now (another tab): this one ended there;
        // the added servers' tokens are that login's.
        hub.detach();
        account = null;
        return leave("Another login is active in this browser now.", false);
      }
      if (ok) account = null;
      leave(ok ? null : "Logging out did not finish: you may still be logged in.");
    },
    logoutAll: async () => {
      void stopPush();
      const mine = authGeneration;
      await api.logoutAll();
      if (mine !== authGeneration) return;
      account = null;
      leave("Logged out on all devices.");
    },
    changePassword: async (old, next) => {
      await api.changePassword(old, next);
      if (account) account = { ...account, temporary: false, password: true }; // a guest account is kept now
    },
    get hasPassword() {
      return account?.password !== false;
    },
    get fedi() {
      return account?.fedi ?? null;
    },
    fediEnabled: () => api.fediEnabled(),
    fediLink: async (instance, password) => {
      location.assign(await api.fediStart(instance, password)); // back with ?fedi=linked
    },
    fediUnlink: async () => {
      await api.fediUnlink();
      if (account) account = { ...account, fedi: null };
    },
    fediPublic: async (on) => {
      await api.fediPublic(on);
      if (account?.fedi) account = { ...account, fedi: { ...account.fedi, public: on } };
    },
    fediReauth: async () => {
      if (!account?.fedi) return;
      location.assign(await api.fediReauth(reauthInstance(account.fedi.handle, safeGet(localStorage, "fediInstance")))); // back with ?fedi=reauth
    },
    get whipUrl() {
      return api.whipUrl;
    },
    streamKey: () => api.streamKey(),
    newStreamKey: () => api.newStreamKey(),
    removeStreamKey: () => api.removeStreamKey(),
  };

  // The desktop app: a Kallisti server works as in the web app (with a
  // session token instead of the cookie); any other address through the
  // app's local Mumble bridge. The web app: through its Kallisti server.
  const desktop = isDesktop();
  let desktopServer = $state(safeGet(localStorage, "server") ?? "");
  let desktopKind = $state<"kallisti" | "mumble" | null>(null);
  let kallisti = $state<string | null>(null); // the Kallisti server's origin
  const sessionStore = (origin: string) => {
    const key = `session:${origin}`;
    return { get: () => safeGet(localStorage, key) || null, set: (token: string | null) => void safeSet(localStorage, key, token ?? "") };
  };
  // The newest choice wins (the one from last time is looked at on start).
  let choices = 0;
  const chooseServer = async (address: string) => {
    const mine = ++choices;
    loginNote = null;
    const origin = kallistiOrigin(address);
    const kind = origin === null ? "other" : await isKallisti(origin);
    if (mine !== choices) return;
    if (kind === "unreachable") {
      loginNote = `Cannot reach ${address}. Check the address, or try again.`;
      checking = false;
      return;
    }
    const found = kind === "kallisti";
    desktopServer = address;
    safeSet(localStorage, "server", address);
    kallisti = found ? origin : null;
    if (found) {
      api = createAccountsApi(origin!, fetch, sessionStore(origin!));
      const generation = authGeneration;
      const answer = await api.check();
      if (mine !== choices || generation !== authGeneration) return;
      desktopKind = "kallisti";
      if (answer !== "unknown") ownerKnown = true;
      return autoLogin(answer === "unknown" ? null : answer);
    }
    // A Mumble server (or a Kallisti server that failed: the app cannot
    // tell): not a sure answer about the home login (see ownerKnown).
    desktopKind = "mumble";
    checking = false;
  };
  const changeServer = () => {
    authGeneration++;
    hub.endAll(); // another home server: not with this one's added servers
    choices++;
    desktopKind = null;
    kallisti = null;
    account = null;
  };
  // A server from last time: looked at again at once.
  untrack(() => desktop && (desktopServer ? void chooseServer(desktopServer) : (checking = false)));

  // local: through the desktop app's own bridge (a Mumble server).
  const urlFor = async (credentials: Credentials): Promise<Target> => {
    if (!desktop) return { url: PROXY_URL };
    if (kallisti) return kallistiSocket(kallisti, sessionStore(kallisti).get());
    const server = parseAddress(credentials.server ?? "");
    if (!server) throw new Error("Enter a server address, for example voice.example.org or voice.example.org:64738");
    return { url: bridgeUrl(await getBridge(), server), local: true };
  };
  // This server's connection, its reconnects and what goes with it (see
  // lib/serverSession).
  const session = createServerSession<Connection>({
    site: () => kallisti ?? location.origin,
    urlFor,
    connect: (target, c) => connect(target.url, c, audio, target.protocols),
    tabStorage,
  });
  $effect(() => session.dispose);
  const connection = $derived(session.connection);

  // Several servers in one client (see lib/serverHub): the home account's
  // list, each server with its own token on this device and its session;
  // all stay connected, one is on screen (null: the home server).
  const homeOrigin = () => kallisti ?? location.origin;
  const memoryStore = (token: string | null = null) => ({ get: () => token, set: (t: string | null) => void (token = t) });
  const hub = createServerHub<ServerSession<Connection>>({
    storage: () => {
      try {
        return localStorage;
      } catch {
        return undefined;
      }
    },
    allowed: (origin) => {
      const r = serverOrigin(origin, homeOrigin(), !desktop);
      return "origin" in r && r.origin === origin;
    },
    // Each call names the account this page shows (a stale window after
    // another login in this browser is refused).
    list: () => {
      const name = owner === null ? null : account?.name;
      return name ? { load: () => api.servers(name), add: (s) => api.addServer(s, name), remove: (o) => api.removeServer(o, name) } : null;
    },
    owner: () => owner,
    open: (origin, name, token) => {
      const s = createServerSession<Connection>({
        site: () => origin,
        urlFor: async () => kallistiSocket(origin, token),
        connect: (target, c) => connect(target.url, c, audio, target.protocols),
        tabStorage,
      });
      void s.start({ username: name, account: true }).catch(() => {});
      return s;
    },
    revoke: async (origin, token) => void (await createAccountsApi(origin, fetch, memoryStore(token)).logout()),
  });
  // Whose the added servers' tokens are: an account at home (not a guest,
  // not a Mumble server in the desktop app). Once the first check is done,
  // every change goes to the hub (another login ends the previous one's).
  const owner = $derived(account && !account.temporary ? `${homeOrigin()} ${account.name}` : null);
  $effect(() => {
    const now = owner;
    if (ownerKnown) untrack(() => hub.follow(now));
  });
  // The add dialog, for the home login that opened it (another login
  // closes it, and a login there that ends later is not taken).
  // server: one of the list that needs a login on this device.
  let addingFor = $state<{ owner: string; server?: { origin: string; name: string }; address?: string } | null>(null);
  const addingServer = $derived(addingFor !== null && addingFor.owner === owner);
  // Closed by another login: forgotten (it must not open again later).
  $effect(() => {
    if (addingFor !== null && addingFor.owner !== owner) addingFor = null;
  });
  // The add dialog's logins: the token stays in memory until the hub
  // takes it.
  const dialogApi = (origin: string, token?: string) => {
    const store = memoryStore(token ?? null);
    return { api: createAccountsApi(origin, fetch, store), token: store.get };
  };
  // Once the home server is in: the saved ones (once per home login).
  $effect(() => {
    if (connection?.state.status === "synced") untrack(() => void hub.sync());
  });
  // Changes from other devices and tabs: when the page shows again.
  $effect(() => {
    const onShow = () => document.visibilityState === "visible" && void hub.sync();
    document.addEventListener("visibilitychange", onShow);
    return () => document.removeEventListener("visibilitychange", onShow);
  });
  $effect(() => () => hub.close());
  const allSessions = () => [session, ...hub.servers.flatMap((a) => (a.session ? [a.session] : []))];
  const shownAdded = $derived(hub.shown === null ? null : (hub.servers.find((a) => a.origin === hub.shown) ?? null));
  // Each server's mark (see Client onBadge).
  let badges = $state<Record<string, RailBadge>>({});
  const setBadge = (id: string, b: RailBadge) => {
    const old = untrack(() => badges[id]);
    if (old?.mentions !== b.mentions || old.unread !== b.unread) badges = { ...untrack(() => badges), [id]: b };
  };
  const railEntry = (id: string, c: Connection | null, fallback: string): RailServer => ({
    ...badges[id],
    id,
    name: c?.state.channels[0]?.name ?? fallback,
    logo: c?.text.logo ? iconHref(c.httpBase, 192, c.text.logo) : null,
    active: hub.shown === (id || null),
  });
  const railServers = $derived<RailServer[]>([
    railEntry("", connection, "Home"),
    ...hub.servers.map((a) => railEntry(a.origin, a.session?.connection ?? null, new URL(a.origin).host)),
  ]);
  // A dialog open in the client that goes out of sight would keep the page
  // inert: it closes (a notification's click switches servers).
  const selectServer = (id: string) => {
    if ((id || null) !== hub.shown) for (const d of document.querySelectorAll<HTMLDialogElement>(".view dialog[open]")) d.close();
    hub.show(id || null);
  };

  // Push from added servers (see lib/push syncServerPush): while the home
  // server's is on (in this browser, for this login).
  let homePushOn = $state(false);
  $effect(() => {
    const name = owner !== null ? account?.name : undefined;
    const check = () => {
      if (name && browserPushSupport() === "ok") void pushIsOn(httpBase(PROXY_URL), name).then((on) => (homePushOn = on));
      else homePushOn = false;
    };
    check();
    addEventListener("kallisti-push", check);
    return () => removeEventListener("kallisti-push", check);
  });
  const pushSyncFor = (a: HubServer<ServerSession<Connection>>) => (prefs: PushPrefs) => {
    if (a.token) void syncServerPush({ origin: a.origin, token: a.token, name: a.name }, homePushOn, prefs);
  };
  // Servers that went (removed, logged out): their registrations too.
  $effect(() => {
    const keep = hub.servers.filter((s) => s.token).map((s) => s.origin);
    if (browserPushSupport() === "ok") untrack(() => void dropServerPushes(keep));
  });

  // Invites (see lib/invites): a link to a server, clicked in a client or
  // opened here with ?join=<link>: that server on screen (added first, with
  // the home account or a password) and the invite redeemed there (its
  // client takes the link, see Client link).
  let pendingLinks = $state<Record<string, string>>({});
  // Another home login: none of the last one's.
  let linksOf = untrack(() => owner);
  $effect(() => {
    if (owner === linksOf) return;
    linksOf = owner;
    untrack(() => (pendingLinks = {}));
  });
  const openInvite = (url: string): boolean => {
    const inv = inviteOf(url);
    if (!inv) return false;
    const known = hub.servers.find((s) => s.origin === inv.origin);
    const route = inviteRoute(inv.origin, new URL(homeOrigin()).origin, known && { session: !!known.session }, owner === null);
    if (route === null) return false;
    const id = route === "home" ? "" : inv.origin;
    pendingLinks[id] = url;
    if (route !== "add") selectServer(id);
    if (route === "login") addingFor = { owner: owner!, server: { origin: known!.origin, name: known!.name } };
    if (route === "add") addingFor = { owner: owner!, address: inv.origin };
    return true;
  };
  const linkUsed = (id: string) => () => delete pendingLinks[id];
  // A message link of a server here (a push notice's click, ?open=).
  const openMessageLink = (url: string) => {
    let origin: string;
    try {
      origin = new URL(url).origin;
    } catch {
      return;
    }
    const id = origin === new URL(homeOrigin()).origin ? "" : hub.servers.find((s) => s.origin === origin && s.session)?.origin;
    if (id === undefined) return;
    pendingLinks[id] = url;
    selectServer(id);
  };
  let openUrl = new URLSearchParams(location.search).get("open");
  if (openUrl) history.replaceState(history.state, "", location.pathname);
  // Once the added servers are there (restored after the home connects).
  $effect(() => {
    if (openUrl && hub.servers.some((s) => openUrl!.startsWith(`${s.origin}/`) && s.session?.connection))
      untrack(() => {
        openMessageLink(openUrl!);
        openUrl = null;
      });
  });
  // ?join=<an invite link> (another server's invite page sends people to
  // their home): once the home login is known and connected.
  let joinUrl = new URLSearchParams(location.search).get("join");
  if (joinUrl) history.replaceState(history.state, "", location.pathname);
  // Kept over a login through another server (see keptInvite).
  if (joinUrl) safeSet(tabStorage(), "kallisti-invite", `/?join=${encodeURIComponent(joinUrl)}`);
  $effect(() => {
    if (joinUrl && connection?.state.status === "synced" && ownerKnown && (hub.listed || owner === null))
      untrack(() => {
        safeSet(tabStorage(), "kallisti-invite", "");
        if (!openInvite(joinUrl!)) fediNotice = inviteOf(joinUrl!) ? "Log in with an account (not as a guest) to join other servers." : "This is not an invite link.";
        joinUrl = null;
      });
  });

  // Voice across servers: the clients of all connected servers stay
  // mounted (one shown), so voice goes on while another is on screen. One
  // voice channel at a time: joining one elsewhere leaves the other.
  const views = $derived([{ id: "", session }, ...hub.servers.flatMap((a) => (a.session ? [{ id: a.origin, session: a.session }] : []))]);
  // A lost or closed connection's voice is not one (nothing goes through it).
  const inVoiceOn = (s: ServerSession<Connection>) =>
    !s.lost && !s.error && !s.connection?.closed && s.connection?.state.status === "synced" && inVoice(s.connection.state);
  const voiceId = $derived(views.find((v) => inVoiceOn(v.session))?.id ?? null);
  const leaveVoiceOn = (id: string) => {
    const c = views.find((v) => v.id === id)?.session.connection;
    const me = c?.state.session == null ? undefined : c.state.users[c.state.session];
    if (c && me && isLobby(c.state)) c.send(joinChannel(me.session, 0));
  };
  let voiceBefore: string[] = [];
  $effect(() => {
    const ids = views.filter((v) => inVoiceOn(v.session)).map((v) => v.id);
    untrack(() => {
      const stays = voiceJoined(voiceBefore, ids);
      voiceBefore = ids;
      if (stays === null) return;
      // The others leave, and do not rejoin after a reconnect.
      for (const v of views.filter((x) => x.id !== stays)) {
        if (ids.includes(v.id)) leaveVoiceOn(v.id);
        const place = v.session.place;
        if (place?.voice != null) v.session.keepPlace({ ...place, voice: null });
      }
    });
  });
  // The voice bar of the server in voice, for the others' screens.
  const remoteVoiceFor = (id: string): RemoteVoice | null => {
    const v = voiceId === null || voiceId === id ? null : views.find((x) => x.id === voiceId);
    const c = v?.session.connection;
    const me = c?.state.session == null ? undefined : c.state.users[c.state.session];
    if (!v || !c || !me) return null;
    const target = v.id;
    return {
      key: target || (kallisti ?? location.origin),
      server: railServers.find((r) => r.id === target)?.name ?? "",
      channel: c.state.channels[me.channelId]?.name ?? "",
      selfMute: me.selfMute,
      selfDeaf: me.selfDeaf,
      info: c.info,
      setSelf: (s) => c.send(setSelfState(me.session, s)),
      ptt: (down) => c.capture?.setPttDown(down),
      talking: c.capture?.transmitting ?? false,
      leave: () => leaveVoiceOn(target),
      show: () => selectServer(target),
    };
  };
  // Push-to-talk acts on the server in voice, else on the one shown.
  const pttOn = (id: string) => (voiceId ?? (hub.shown ?? "")) === id;
  // An added server's account in its settings: logging out takes it out
  // of the rail; no fediverse link from here (that is its own site's).
  const addedActions = (origin: string): AccountActions => {
    // Its own token (another tab may store another login's for the origin).
    const api = createAccountsApi(origin, fetch, memoryStore(hub.servers.find((a) => a.origin === origin)?.token ?? null));
    return {
      temporary: false,
      logout: async () => hub.remove(origin),
      logoutAll: async () => {
        // The entry this was for (another may have taken its place meanwhile).
        const entry = hub.servers.find((a) => a.origin === origin);
        await api.logoutAll();
        if (entry && hub.servers.includes(entry)) hub.remove(origin);
      },
      changePassword: (old, next) => api.changePassword(old, next),
      whipUrl: api.whipUrl,
      streamKey: () => api.streamKey(),
      newStreamKey: () => api.newStreamKey(),
      removeStreamKey: () => api.removeStreamKey(),
      hasPassword: true,
      fedi: null,
      fediEnabled: async () => false,
      fediLink: async () => {},
      fediUnlink: async () => {},
      fediPublic: async () => {},
      fediReauth: async () => {},
    };
  };

  // A new version on the server (web app only: the desktop app carries its
  // own): checked now, every 15 minutes, when the tab shows again, and
  // after each (re)connect, as a deploy restarts the server. No automatic
  // reload: it would end voice.
  let updateReady = $state(false);
  let updateClosed = $state(false);
  const checkUpdate = async () => {
    if (!desktop && !updateReady && document.visibilityState === "visible") updateReady = await checkForUpdate(location.origin);
  };
  $effect(() => {
    void checkUpdate();
    const timer = setInterval(checkUpdate, 15 * 60_000);
    const onShow = () => void checkUpdate();
    document.addEventListener("visibilitychange", onShow);
    return () => {
      clearInterval(timer);
      document.removeEventListener("visibilitychange", onShow);
    };
  });
  $effect(() => {
    if (connection?.state.status === "synced") untrack(() => void checkUpdate());
  });

  // Desktop: the server's certificate is new or changed. The user decides.
  const trustAndConnect = async () => {
    const certificate = session.certificate;
    const server = parseAddress(certificate?.server ?? "");
    const credentials = session.credentials;
    if (!certificate || !server || !credentials) return;
    try {
      await trustServer(server, certificate.fingerprint);
    } catch (err) {
      session.drop();
      loginNote = `Could not trust the certificate: ${err instanceof Error ? err.message : String(err)}`;
      return;
    }
    await start(credentials);
  };

  // The audio context lives as long as the user stays: reconnects use it
  // again (a new one would need another click).
  let audio: AudioContext | null = null;
  // Started without a click (a saved login, a refresh), browsers may hold
  // it suspended: the first click or key turns the sound on. Not in the
  // Linux app: it holds the context itself for a while (see startAfter).
  let audioState = $state<AudioContextState | null>(null);
  const wakeSound = () => {
    if (audioState === "suspended") void audio?.resume().catch(() => {});
  };
  // Made in a click where possible: browsers allow audio only after a
  // user gesture.
  const makeAudio = () => {
    const ctx = new AudioContext({ sampleRate: 48_000, latencyHint: "interactive" });
    // Linux app: the capture module first, then the output (see startAfter).
    if (linuxApp()) startAfter(ctx, loadCaptureWorklet(ctx));
    audioState = linuxApp() ? null : ctx.state;
    if (!linuxApp()) ctx.onstatechange = () => audio === ctx && (audioState = ctx.state);
    return ctx;
  };
  const closeAudio = (ctx: AudioContext | null) => {
    if (ctx) closeVoiceOutput(ctx);
    void ctx?.close().catch(() => {});
  };
  // Into voice, or a new output device: everything moves to a fresh
  // context (see lib/voice/restart). Not on reconnects. keepMic: the
  // microphone stays open (a new output device).
  const restartAudio = (keepMic = false) => {
    const move = (ctx: AudioContext) => allSessions().forEach((s) => s.setAudio(ctx, keepMic));
    audio = swapContext(audio, makeAudio, move, closeAudio);
  };
  const wakeAll = () => allSessions().forEach((s) => s.wake());
  const start = async (credentials: Credentials) => {
    loginNote = null;
    // The old connection closes at once (before the address is known).
    const started = session.start(credentials);
    closeAudio(audio);
    // Created here: a login is a user gesture.
    audio = voiceSupported() ? makeAudio() : null;
    if (!audio) audioState = null;
    else for (const a of hub.servers) a.session?.setAudio(audio, true); // their microphones stay
    try {
      await started;
    } catch (err) {
      loginNote = err instanceof Error ? err.message : String(err);
    }
  };
</script>

<svelte:window ononline={wakeAll} onpagehide={() => allSessions().forEach((s) => (s.place && s.keepPlace(s.place, true), s.keepStreams()))} onpointerdowncapture={wakeSound} onkeydowncapture={wakeSound} />
<svelte:document onvisibilitychange={wakeAll} />

<!-- The rail of servers: the same while a server connects, fails or
     reconnects (in the client's drawer on phones; beside these screens). -->
{#snippet rail()}
  <ServerRail servers={railServers} onSelect={selectServer} onAdd={owner !== null ? () => (addingFor = { owner: owner! }) : undefined} />
{/snippet}
<!-- A screen without the client, with the rail if there are servers to switch to. -->
{#snippet withRail(content: Snippet)}
  {#if hub.servers.length}
    <div class="shell">{@render rail()}<div class="shell-main">{@render content()}</div></div>
  {:else}
    {@render content()}
  {/if}
{/snippet}

{#if connectHome && connectPhase}
  <Connect {api} home={connectHome} phase={connectPhase} />
{:else if authorizeQuery && account && !checking}
  <Authorize {api} query={authorizeQuery} serverName={location.host} />
{:else if checking}
  <div class="connecting" aria-busy="true"></div>
{:else}
  <!-- Every connected server's client, the one on screen visible. -->
  {#if connection?.state.status === "synced" && !session.error}
    {@const shown = hub.shown === null}
    <div class="view" hidden={!shown} inert={!shown}>
      <!-- Quiet first (the server is usually back within seconds): a small
           sign, then the banner (see lossNotice). -->
      {#if shown && session.lost && session.notice === "banner"}
        <div class="reconnecting" role="status">
          Connection lost. Reconnecting…
          <button type="button" onclick={() => void session.tryNow()}>Retry now</button>
        </div>
      {:else if shown && session.lost && session.notice === "sign"}
        <div class="reconnecting quiet" role="status">Reconnecting…</div>
      {/if}
      <!-- While lost, the old view stays readable but takes no input (it would be lost); the rail does. -->
      {#key connection}
        <Client {connection} link={pendingLinks[""] ?? null} onLinkUsed={linkUsed("")} onInvite={openInvite} onBadge={(b) => setBadge("", b)} active={shown} pttHere={pttOn("")} remoteVoice={remoteVoiceFor("")} onShow={() => selectServer("")} lost={session.lost} {rail} account={account ? accountActions : undefined} resume={session.place} onPlace={(p) => session.keepPlace(p)} {audioState} onRestartAudio={restartAudio} onLogIn={() => leave(null)} site={kallisti ?? location.origin} drafts={session.drafts} pendingFiles={session.pendingFiles} streams={session.streams} openSettings={reopenSettings} onSettingsOpened={() => (reopenSettings = false)} />
      {/key}
    </div>
  {/if}
  {#each hub.servers as a (a.origin)}
    {@const s = a.session}
    {#if s?.connection?.state.status === "synced" && !s.error}
      {@const shown = hub.shown === a.origin}
      <div class="view" hidden={!shown} inert={!shown}>
        {#if shown && s.lost}<div class="reconnecting quiet" role="status">Reconnecting…</div>{/if}
        {#key s.connection}
          <Client connection={s.connection} pushSync={pushSyncFor(a)} link={pendingLinks[a.origin] ?? null} onLinkUsed={linkUsed(a.origin)} onInvite={openInvite} onBadge={(b) => setBadge(a.origin, b)} active={shown} pttHere={pttOn(a.origin)} remoteVoice={remoteVoiceFor(a.origin)} onShow={() => selectServer(a.origin)} lost={s.lost} {rail} account={addedActions(a.origin)} push={false} resume={s.place} onPlace={(p) => s.keepPlace(p)} {audioState} onRestartAudio={restartAudio} site={a.origin} drafts={s.drafts} pendingFiles={s.pendingFiles} streams={s.streams} />
        {/key}
      </div>
    {/if}
  {/each}
  {#if shownAdded}
    {@const a = shownAdded}
    {@const s = a.session}
    {#if !(s?.connection?.state.status === "synced" && !s.error)}
      {#snippet added()}
        {@const host = new URL(a.origin).host}
        <div class="connecting">
          <p>{!s ? `Log in to ${host} on this device.` : s.error ? `${host}: ${s.error}` : `Connecting to ${host}…`}</p>
          <p class="actions">
            {#if !s && owner !== null}
              <button type="button" onclick={() => (addingFor = { owner: owner!, server: { origin: a.origin, name: a.name } })}>Log in</button>
            {:else if s?.error}
              <button type="button" onclick={() => void s.retry().catch(() => {})}>Try again</button>
            {/if}
            <button type="button" onclick={() => hub.remove(a.origin)}>Remove this server</button>
          </p>
        </div>
      {/snippet}
      {@render withRail(added)}
    {/if}
  {:else if !connection || session.error}
    {#snippet login()}
      <Login
        error={session.certificate ? null : (session.error ?? loginNote)}
        certificate={session.certificate}
        onTrust={trustAndConnect}
        onCancelTrust={() => leave(null, false)}
        {api}
        {account}
        desktop={desktop ? { server: desktopServer, kind: desktopKind, onServer: chooseServer, onChangeServer: changeServer } : null}
        onAccountChange={(acc) => {
          // A logout from the login screen (after a kick or refusal) ends the
          // home login too.
          authGeneration++;
          if (!acc) hub.endAll();
          account = acc;
          ownerKnown = true;
        }}
        onConnect={start}
        {fediSignup}
        onFediSignupDone={() => (connectHome ? window.close() : (fediSignup = false))}
        onHandoff={() => (connectPhase = "deliver")}
      />
    {/snippet}
    {@render withRail(login)}
  {:else if connection.state.status !== "synced"}
    {#snippet connectingHome()}
      <div class="connecting">Connecting…</div>
    {/snippet}
    {@render withRail(connectingHome)}
  {/if}
{/if}

<div class="status-bar" aria-hidden="true"></div>

{#if fediNotice}<Toast text={fediNotice} />{/if}
{#key hub.error}{#if hub.error}<Toast text={hub.error} />{/if}{/key}

{#if addingServer}
  <AddServerDialog
    home={homeOrigin()}
    homeCookie={!desktop}
    taken={hub.servers.map((a) => a.origin)}
    apiFor={dialogApi}
    server={addingFor?.server}
    address={addingFor?.address}
    onAdded={(origin, acc, token) => addingFor?.owner === owner && (addingFor.server ? hub.login(origin, acc.name, token) : hub.add(origin, acc.name, token))}
    onClose={() => {
      // An invite for a server that did not come: forgotten.
      const origin = addingFor?.server?.origin ?? addingFor?.address;
      if (origin && !hub.servers.some((s) => s.origin === origin && s.session)) delete pendingLinks[origin];
      addingFor = null;
    }}
  />
{/if}

{#if updateReady && !updateClosed}
  <div class="update" role="status">
    A new version of Kallisti is available.
    <button type="button" onclick={() => location.reload()}>Reload</button>
    <button type="button" class="close" aria-label="Close" onclick={() => (updateClosed = true)}>×</button>
  </div>
{/if}

<style>
  /* The installed iOS app (iOS 26 and later) blurs the page below the
     status bar unless a fixed box at least 90% wide, over 10px tall, with
     one background colour sits at the top (WebKit's
     LocalFrameView::fixedContainerEdges); then the status bar gets that
     colour. Only with iOS's own status bar, not black-translucent.
     background-clip: text keeps the box itself invisible. */
  .status-bar {
    display: none;
  }
  @supports (-webkit-touch-callout: none) {
    @media (display-mode: standalone) {
      .status-bar {
        display: block;
        position: fixed;
        inset: 0 0 auto;
        z-index: 40;
        height: 11px;
        background-color: var(--bg-main);
        background-clip: text;
        pointer-events: none;
      }
    }
  }
  .update {
    position: fixed;
    top: calc(8px + var(--safe-top));
    left: 50%;
    transform: translateX(-50%);
    z-index: 60;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 8px 8px 14px;
    border-radius: 8px;
    background: var(--bg-menu);
    box-shadow: 0 8px 24px rgb(0 0 0 / 0.4);
    font-size: 14px;
    white-space: nowrap;
  }
  .update button {
    padding: 4px 12px;
    border-radius: 4px;
    background: var(--accent-strong);
    color: #fff;
    font-weight: 600;
  }
  .update .close {
    padding: 0 6px;
    background: none;
    color: var(--text-muted);
    font-size: 18px;
  }
  .reconnecting {
    position: fixed;
    top: calc(8px + var(--safe-top));
    left: 50%;
    transform: translateX(-50%);
    z-index: 50;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 6px 8px 6px 14px;
    border: 1px solid var(--danger);
    border-radius: 6px;
    background: var(--bg-rail);
    color: var(--text);
    font-size: 13px;
    font-weight: 600;
    box-shadow: 0 2px 8px rgb(0 0 0 / 0.3);
  }
  .reconnecting.quiet {
    padding: 3px 10px;
    border-color: var(--border-subtle);
    color: var(--text-muted);
    font-size: 12px;
    font-weight: 400;
    box-shadow: none;
    opacity: 0.85;
  }
  .reconnecting button {
    padding: 4px 10px;
    border-radius: 4px;
    background: var(--bg-active);
    color: var(--text);
  }
  /* The clients of servers not on screen stay mounted, out of sight but
     with their size (display: none would lose the lists' scroll position,
     and nothing could measure itself). */
  .view:not([hidden]) {
    display: contents;
  }
  .view[hidden] {
    display: block;
    position: fixed;
    inset: 0;
    z-index: -1;
    visibility: hidden;
    pointer-events: none;
  }
  /* A screen without the client, beside the rail. */
  .shell {
    display: grid;
    grid-template-columns: calc(72px + var(--safe-left)) 1fr;
    height: 100%;
  }
  .shell > :global(nav) {
    border-top: var(--safe-top) solid transparent;
    border-bottom: var(--safe-bottom) solid transparent;
    border-left: var(--safe-left) solid transparent;
  }
  .shell-main {
    min-width: 0;
    height: 100%;
    overflow: auto;
  }
  .connecting {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    height: 100%;
    padding: 16px;
    color: var(--text-muted);
    text-align: center;
  }
  .connecting p {
    margin: 0;
  }
  .connecting .actions {
    display: flex;
    gap: 8px;
  }
  .connecting .actions button {
    padding: 6px 12px;
    border-radius: 4px;
    background: var(--bg-input);
    color: var(--text);
  }
</style>
