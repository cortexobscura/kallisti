<script lang="ts">
  import Client from "./components/Client.svelte";
  import Login from "./components/Login.svelte";
  import AddServerDialog from "./components/AddServerDialog.svelte";
  import type { RailServer } from "./components/ServerRail.svelte";
  import { createAccountsApi, type Account, type AccountActions } from "./lib/accounts";
  import { connect, type Connection, type Credentials } from "./lib/connection.svelte";
  import { createServerSession, type ServerSession, type Target } from "./lib/serverSession.svelte";
  import { createServerHub } from "./lib/serverHub.svelte";
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
  import { safeGet, safeSet } from "./lib/storage";
  import { proxyUrl } from "./lib/status";
  import { loggedOut } from "./lib/reconnect";
  import { untrack } from "svelte";
  import { checkForUpdate } from "./lib/update";
  import { browserPushSupport, turnOffPush } from "./lib/push";
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
    if (!acc || connection || account !== null) return void (checking = false);
    account = acc;
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
  if (fediBack) history.replaceState(history.state, "", withoutFediResult(location.pathname + location.search + location.hash));
  let fediSignup = $state(fediBack?.kind === "register");
  let fediNotice = $state<string | null>(null);
  // Back from a fresh login for the first password: to the settings again.
  let reopenSettings = $state(fediBack?.kind === "reauth");
  if (!isDesktop())
    void untrack(() => api)
      .check()
      .then((answer) => {
        // Unknown (the server or the network failed): the login form, but
        // nothing of the last login ends (see ownerKnown).
        if (answer !== "unknown") ownerKnown = true;
        const acc = answer === "unknown" ? null : answer;
        // Logged in: a notice over the app (and no name step); else errors
        // on the login screen.
        if (acc) fediSignup = false;
        if (fediBack && fediBack.kind !== "register") {
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
  // another device): back to the login screen, as after a logout.
  $effect(() => {
    if (connection?.closed && loggedOut(connection.closeReason))
      untrack(() => {
        account = null;
        void stopPush();
        leave("You were logged out.");
      });
  });
  // This browser's push notifications end with the login (see lib/push).
  const stopPush = () => (browserPushSupport() === "ok" ? turnOffPush(httpBase(PROXY_URL)) : Promise.resolve());
  const accountActions: AccountActions = {
    get temporary() {
      return account?.temporary ?? false;
    },
    logout: async () => {
      void stopPush(); // not waited for: the logout ends it on the server too
      const ok = await api.logout();
      if (ok) account = null;
      leave(ok ? null : "Logging out did not finish: you may still be logged in.");
    },
    logoutAll: async () => {
      void stopPush();
      await api.logoutAll();
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
  const tabStorage = () => {
    // Reading sessionStorage can throw (blocked storage).
    try {
      return sessionStorage;
    } catch {
      return undefined;
    }
  };
  const session = createServerSession<Connection>({
    site: () => kallisti ?? location.origin,
    urlFor,
    connect: (target, c) => connect(target.url, c, audio, target.protocols),
    tabStorage,
  });
  $effect(() => session.dispose);
  const connection = $derived(session.connection);

  // Several servers in one client (see lib/serverHub): the servers added
  // on this device, each with its own token and session; all stay
  // connected, one is on screen (null: the home server). Their tokens have
  // keys of their own (in the desktop app the home server's is session:).
  const homeOrigin = () => kallisti ?? location.origin;
  const addedStore = (origin: string) => {
    const key = `added-session:${origin}`;
    return { get: () => safeGet(localStorage, key) || null, set: (token: string | null) => void safeSet(localStorage, key, token ?? "") };
  };
  const memoryStore = (token: string | null = null) => ({ get: () => token, set: (t: string | null) => void (token = t) });
  const hub = createServerHub<ServerSession<Connection>>({
    storage: () => {
      try {
        return localStorage;
      } catch {
        return undefined;
      }
    },
    home: homeOrigin,
    owner: () => owner,
    tokens: addedStore,
    open: (origin, name, token) => {
      const s = createServerSession<Connection>({
        site: () => origin,
        // Its own token (another tab may store another login's).
        urlFor: async () => kallistiSocket(origin, token),
        connect: (target, c) => connect(target.url, c, audio, target.protocols),
        tabStorage,
      });
      void s.start({ username: name, account: true }).catch(() => {});
      return s;
    },
    revoke: async (origin, token) => void (await createAccountsApi(origin, fetch, memoryStore(token)).logout()),
  });
  // Whose the added servers are: an account at home (not a guest, not a
  // Mumble server in the desktop app). Once the first check is done, every
  // change goes to the hub (another login ends the previous one's).
  const owner = $derived(account && !account.temporary ? `${homeOrigin()} ${account.name}` : null);
  $effect(() => {
    const now = owner;
    if (ownerKnown) untrack(() => hub.follow(now));
  });
  // The add dialog, for the home login that opened it (another login
  // closes it, and a login there that ends later is not taken).
  let addingFor = $state<string | null>(null);
  const addingServer = $derived(addingFor !== null && addingFor === owner);
  // Closed by another login: forgotten (it must not open again later).
  $effect(() => {
    if (addingFor !== null && addingFor !== owner) addingFor = null;
  });
  // The add dialog's logins: the token stays in memory until the hub
  // takes it.
  const dialogApi = (origin: string) => {
    const store = memoryStore();
    return { api: createAccountsApi(origin, fetch, store), token: store.get };
  };
  // Once the home server is in: the saved ones (once per home login).
  $effect(() => {
    if (connection?.state.status === "synced") untrack(() => hub.restore());
  });
  $effect(() => () => hub.close());
  const allSessions = () => [session, ...hub.servers.map((a) => a.session)];
  const shownAdded = $derived(hub.shown === null ? null : (hub.servers.find((a) => a.origin === hub.shown) ?? null));
  const railEntry = (id: string, c: Connection | null, fallback: string): RailServer => ({
    id,
    name: c?.state.channels[0]?.name ?? fallback,
    logo: c?.text.logo ? iconHref(c.httpBase, 192, c.text.logo) : null,
    active: hub.shown === (id || null),
  });
  const rail = $derived<RailServer[]>([
    railEntry("", connection, "Home"),
    ...hub.servers.map((a) => railEntry(a.origin, a.session.connection, new URL(a.origin).host)),
  ]);
  const selectServer = (id: string) => hub.show(id || null);
  const canAdd = $derived(owner !== null);
  // An added server's account in its settings: logging out takes it out
  // of the rail; no fediverse link from here (that is its own site's).
  const addedActions = (origin: string): AccountActions => {
    // Its own token (another tab may store another login's for the origin).
    const api = createAccountsApi(origin, fetch, memoryStore(hub.servers.find((a) => a.origin === origin)?.token ?? null));
    return {
      temporary: false,
      logout: async () => hub.remove(origin),
      logoutAll: async () => {
        await api.logoutAll();
        hub.remove(origin);
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
    else for (const a of hub.servers) a.session.setAudio(audio, true); // their microphones stay
    try {
      await started;
    } catch (err) {
      loginNote = err instanceof Error ? err.message : String(err);
    }
  };
</script>

<svelte:window ononline={wakeAll} onpagehide={() => allSessions().forEach((s) => (s.place && s.keepPlace(s.place, true), s.keepStreams()))} onpointerdowncapture={wakeSound} onkeydowncapture={wakeSound} />
<svelte:document onvisibilitychange={wakeAll} />

{#if checking}
  <div class="connecting" aria-busy="true"></div>
{:else if !connection || session.error}
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
    onFediSignupDone={() => (fediSignup = false)}
  />
{:else if shownAdded}
  {@const a = shownAdded}
  {#if a.session.connection?.state.status === "synced" && !a.session.error}
    {#if a.session.lost}<div class="reconnecting quiet" role="status">Reconnecting…</div>{/if}
    <div class="client" inert={a.session.lost}>
      {#key a.session.connection}
        <Client connection={a.session.connection} account={addedActions(a.origin)} push={false} resume={a.session.place} onPlace={(p) => a.session.keepPlace(p)} {audioState} onRestartAudio={restartAudio} site={a.origin} drafts={a.session.drafts} pendingFiles={a.session.pendingFiles} streams={a.session.streams} servers={rail} onSelectServer={selectServer} onAddServer={canAdd ? () => (addingFor = owner) : undefined} />
      {/key}
    </div>
  {:else}
    <div class="connecting">
      <p>{a.session.error ? `${new URL(a.origin).host}: ${a.session.error}` : `Connecting to ${new URL(a.origin).host}…`}</p>
      <p class="actions">
        <button type="button" onclick={() => hub.show(null)}>Back</button>
        {#if a.session.error}
          <button type="button" onclick={() => void a.session.retry().catch(() => {})}>Try again</button>
        {/if}
        <button type="button" onclick={() => hub.remove(a.origin)}>Remove this server</button>
      </p>
    </div>
  {/if}
{:else if connection.state.status === "synced"}
  <!-- Quiet first (the server is usually back within seconds): a small
       sign, then the banner (see lossNotice). -->
  {#if session.lost && session.notice === "banner"}
    <div class="reconnecting" role="status">
      Connection lost. Reconnecting…
      <button type="button" onclick={() => void session.tryNow()}>Retry now</button>
    </div>
  {:else if session.lost && session.notice === "sign"}
    <div class="reconnecting quiet" role="status">Reconnecting…</div>
  {/if}
  <!-- While lost, the old view stays readable but takes no input (it would be lost). -->
  <div class="client" inert={session.lost}>
    {#key connection}
      <Client {connection} account={account ? accountActions : undefined} resume={session.place} onPlace={(p) => session.keepPlace(p)} {audioState} onRestartAudio={restartAudio} onLogIn={() => leave(null)} site={kallisti ?? location.origin} drafts={session.drafts} pendingFiles={session.pendingFiles} streams={session.streams} openSettings={reopenSettings} onSettingsOpened={() => (reopenSettings = false)} servers={rail} onSelectServer={selectServer} onAddServer={canAdd ? () => (addingFor = owner) : undefined} />
    {/key}
  </div>
{:else}
  <div class="connecting">Connecting…</div>
{/if}

<div class="status-bar" aria-hidden="true"></div>

{#if fediNotice}<Toast text={fediNotice} />{/if}

{#if addingServer}
  <AddServerDialog
    home={homeOrigin()}
    homeCookie={!desktop}
    taken={hub.servers.map((a) => a.origin)}
    apiFor={dialogApi}
    onAdded={(origin, acc, token) => addingFor === owner && hub.add(origin, acc.name, token)}
    onClose={() => (addingFor = null)}
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
  .client {
    display: contents;
  }
  .connecting {
    display: grid;
    place-items: center;
    height: 100%;
    color: var(--text-muted);
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
