// What a watched stream's client counts (the stream tile's "Stats"):
// pictures and sound that came, drawn pictures, those too late to play,
// pictures skipped to the next keyframe, and lost sound (gaps in the
// sound's timestamps: 20 ms packets).

const SOUND_US = 20_000;
const NEW_START_US = 5_000_000; // a longer jump (or one back): the stream started over

export type StreamTotals = {
  pictures: number;
  sound: number;
  bytes: number;
  drawn: number;
  latePictures: number;
  lateSound: number;
  skipped: number;
  lostSound: number;
};

export const createStreamStats = () => {
  const t: StreamTotals = { pictures: 0, sound: 0, bytes: 0, drawn: 0, latePictures: 0, lateSound: 0, skipped: 0, lostSound: 0 };
  let lastSound: number | null = null;
  return {
    // Arrivals; late: too late to play (sound knows it once decoded).
    picture: (bytes: number) => {
      t.pictures++;
      t.bytes += bytes;
    },
    late: (what: "picture" | "sound") => void (what === "picture" ? t.latePictures++ : t.lateSound++),
    sound: (bytes: number, timestamp: number) => {
      t.sound++;
      t.bytes += bytes;
      const gap = lastSound === null ? SOUND_US : timestamp - lastSound;
      if (gap > SOUND_US && gap < NEW_START_US) t.lostSound += Math.round(gap / SOUND_US) - 1;
      lastSound = timestamp;
    },
    drawn: () => void t.drawn++,
    skipped: () => void t.skipped++,
    totals: (): StreamTotals => ({ ...t }),
  };
};
export type StreamStats = ReturnType<typeof createStreamStats>;

// Rates between two readings `seconds` apart.
// None when no time passed or the totals went back (a new viewer).
export const perSecond = (a: StreamTotals, b: StreamTotals, seconds: number) =>
  seconds > 0 && b.pictures >= a.pictures && b.sound >= a.sound && b.bytes >= a.bytes && b.drawn >= a.drawn
    ? {
        pictures: (b.pictures - a.pictures) / seconds,
        sound: (b.sound - a.sound) / seconds,
        kbps: ((b.bytes - a.bytes) * 8) / seconds / 1000,
        drawn: (b.drawn - a.drawn) / seconds,
      }
    : null;
