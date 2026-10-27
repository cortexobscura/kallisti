import { initialState, type ServerState } from "@kallisti/protocol";

// A connection whose status a test sets (reactive, as the real one).
export const fakeConnection = (url: string) => {
  let state = $state.raw<ServerState>(initialState);
  let closed = $state(false);
  let wasSynced = $state(false);
  let closeReason = $state("");
  return {
    url,
    get state() {
      return state;
    },
    get closed() {
      return closed;
    },
    get closeReason() {
      return closeReason;
    },
    get wasSynced() {
      return wasSynced;
    },
    sync() {
      state = { ...state, status: "synced" };
      wasSynced = true;
    },
    drop(reason = "") {
      closeReason = reason;
      closed = true;
    },
    close() {
      closed = true;
    },
    setAudio() {},
  };
};
export type FakeConnection = ReturnType<typeof fakeConnection>;
