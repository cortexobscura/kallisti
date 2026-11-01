// Dialogs in categories (see CategoryDialog; not lib/categories, the text channels' groups): on desktop a menu beside
// the open one, on phones a list where each opens as its own page.

export type Category<Id extends string = string, G extends string = string> = { id: Id; title: string; group: G };

// Where it opens: the asked category if it is there; else phones at the
// list (null), desktop at the first.
export const startCategory = <Id extends string>(asked: Id | undefined, list: readonly Category<Id>[], phone: boolean): Id | null => {
  if (asked && list.some((c) => c.id === asked)) return asked;
  return phone ? null : (list[0]?.id ?? null);
};

// The menu's groups, in the order of titles; groups without categories go.
export const groupsOf = <Id extends string, G extends string>(list: readonly Category<Id, G>[], titles: Record<G, string>) =>
  (Object.keys(titles) as G[])
    .map((g) => ({ id: g, title: titles[g], categories: list.filter((c) => c.group === g) }))
    .filter((g) => g.categories.length);
