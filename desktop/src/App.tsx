import { listen } from "@tauri-apps/api/event";
import { useEffect, useRef, useState, type FormEvent, type ReactNode } from "react";
import { ActionBar } from "./ActionBar";
import type { ActionId } from "./actions";
import {
  connect,
  type Item,
  type Session,
  type Snapshot,
  type SnapshotRow,
  type Tray,
} from "./api";
import { attentionActions, formatWhen, joinedNav, statusBucket, trayHeading, trayListPhase, type StatusBucket } from "./items";
import { useAppVersion } from "./useAppVersion";
import { snoozeUntil } from "./snooze";

type Selection = { kind: "attention" } | { kind: "tray"; ref: string };

const POLL_MS = 30_000;

const statusFilters = ["all", "pending", "open", "completed"] as const;
type StatusFilter = (typeof statusFilters)[number];

const snoozePresets = [
  ["hour", "One hour"],
  ["tomorrow", "Tomorrow 9:00"],
  ["week", "Next week"],
] as const;

function messageOf(err: unknown): string {
  return err instanceof Error && err.message ? err.message : "request failed";
}

function listKey(selection: Selection): string {
  return selection.kind === "tray" ? selection.ref : "";
}

export default function App() {
  const apiConfig = window.__TRAY_API__;
  if (!apiConfig) {
    return (
      <div className="shell boot">
        <p className="empty">tray is not installed</p>
        <AppVersion />
      </div>
    );
  }
  return <TrayWindow base={apiConfig.base} token={apiConfig.token} />;
}

function TrayWindow({ base, token }: { base: string; token: string }) {
  const [api] = useState(() => connect(base, token));
  const [session, setSession] = useState<Session | null>(null);
  const [snapshot, setSnapshot] = useState<Snapshot | null>(null);
  const [trays, setTrays] = useState<Tray[]>([]);
  const [selection, setSelection] = useState<Selection>({ kind: "attention" });
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [loaded, setLoaded] = useState<{ key: string; items: Item[] }>({
    key: "",
    items: [],
  });
  const [listErrorKey, setListErrorKey] = useState("");
  const [reading, setReading] = useState<Item | null>(null);
  const [readingOpacity, setReadingOpacity] = useState(1);
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState<"decline" | "snooze" | null>(null);
  const [reason, setReason] = useState("");
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [adding, setAdding] = useState(false);
  const [copied, setCopied] = useState(false);
  const [booted, setBooted] = useState(false);
  const listGen = useRef(0);
  const readingGen = useRef(0);
  const acting = useRef(false);

  useEffect(() => {
    let cancelled = false;
    const fail = (err: unknown) => {
      if (!cancelled) setError(messageOf(err));
    };
    api.session().then((next) => { if (!cancelled) setSession(next); }).catch(fail);
    const firstSnapshot = api.snapshot().then((next) => {
      if (!cancelled) setSnapshot(next);
    });
    const firstTrays = api.trays().then((next) => {
      if (!cancelled) setTrays(Array.isArray(next) ? next : []);
    });
    Promise.allSettled([firstSnapshot, firstTrays]).then((results) => {
      if (cancelled) return;
      const failed = results.find((result) => result.status === "rejected");
      if (failed?.status === "rejected") fail(failed.reason);
      // Remote attention lands after the tray list. Refresh before the window leaves loading.
      api.snapshot().then((next) => {
        if (!cancelled) setSnapshot(next);
      }).catch(() => {}).finally(() => {
        if (!cancelled) setBooted(true);
      });
    });

    const timer = setInterval(() => {
      api.snapshot().then((next) => { if (!cancelled) setSnapshot(next); }).catch(fail);
      api.session().then((next) => { if (!cancelled) setSession(next); }).catch(() => {});
      api.trays().then((next) => { if (!cancelled && Array.isArray(next)) setTrays(next); }).catch(() => {});
    }, POLL_MS);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [api]);

  useEffect(() => {
    let unlisten: (() => void) | undefined;
    listen<string>("open-tray", (event) => {
      const ref = event.payload?.trim();
      if (!ref) return;
      setSelection({ kind: "tray", ref });
    })
      .then((stop) => {
        unlisten = stop;
      })
      .catch(() => {});
    return () => unlisten?.();
  }, []);

  useEffect(() => {
    readingGen.current += 1;
    setReading(null);
    setReadingOpacity(1);
    setPending(null);
    setReason("");
    setTitle("");
    setBody("");
    setAdding(false);
    setCopied(false);
    setStatusFilter("all");
  }, [selection]);

  useEffect(() => {
    if (selection.kind !== "tray") {
      listGen.current += 1;
      return;
    }
    let cancelled = false;
    const ref = selection.ref;
    const key = listKey(selection);
    const load = () => {
      const gen = ++listGen.current;
      return api.items(ref, "all").then((next) => {
        if (cancelled || listGen.current !== gen) return;
        setLoaded({ key, items: Array.isArray(next) ? next : [] });
        setListErrorKey("");
      });
    };
    load().catch((err: unknown) => {
      if (cancelled) return;
      setListErrorKey(key);
      setError(messageOf(err));
    });
    const timer = setInterval(() => {
      load().catch((err: unknown) => {
        if (!cancelled) setError(messageOf(err));
      });
    }, POLL_MS);
    return () => {
      cancelled = true;
      listGen.current += 1;
      clearInterval(timer);
    };
  }, [api, selection]);

  const trayKey = listKey(selection);
  const trayPhase = selection.kind === "tray" ? trayListPhase(loaded.key, trayKey, listErrorKey) : "ready";
  const visibleItems = loaded.key === trayKey ? loaded.items : [];
  const selectedTray = selection.kind === "tray" ? trays.find((tray) => tray.ref === selection.ref) : undefined;
  const trayName = selectedTray
    ? (selectedTray.kind === "joined" ? joinedNav(selectedTray).title : selectedTray.name)
    : selection.kind === "tray" ? selection.ref : "Attention";
  const groups = groupTrays(trays);
  const sections = snapshot && Array.isArray(snapshot.sections) ? snapshot.sections : [];
  const buckets = bucketItems(visibleItems);
  const attention = attentionRows(sections);

  function chooseAttention() {
    if (selection.kind === "attention") return;
    setSelection({ kind: "attention" });
  }

  function chooseTray(ref: string) {
    if (selection.kind === "tray" && selection.ref === ref) return;
    setSelection({ kind: "tray", ref });
  }

  function openItem(id: string) {
    const gen = ++readingGen.current;
    setReadingOpacity(0);
    setError(null);
    setPending(null);
    setCopied(false);
    api.item(id).then((item) => {
      if (readingGen.current !== gen) return;
      setReading({ ...item, actions: item.actions ?? [] });
      setReadingOpacity(1);
    }).catch((err: unknown) => {
      if (readingGen.current !== gen) return;
      setReadingOpacity(1);
      setError(messageOf(err));
    });
  }

  async function refreshAfterWrite(current: Selection) {
    const gen = ++listGen.current;
    const key = listKey(current);
    const [snapR, traysR, itemsR] = await Promise.allSettled([
      api.snapshot(),
      api.trays(),
      current.kind === "tray" ? api.items(current.ref, "all") : Promise.resolve(null),
    ]);
    if (snapR.status === "fulfilled") setSnapshot(snapR.value);
    if (traysR.status === "fulfilled") setTrays(Array.isArray(traysR.value) ? traysR.value : []);
    if (current.kind === "tray" && itemsR.status === "fulfilled" && listGen.current === gen) {
      setLoaded({ key, items: Array.isArray(itemsR.value) ? itemsR.value : [] });
    }
    if (snapR.status === "rejected") throw snapR.reason;
    if (current.kind === "tray" && itemsR.status === "rejected") throw itemsR.reason;
  }

  async function runAct(id: string, action: ActionId, body?: unknown) {
    if (acting.current) return;
    acting.current = true;
    setError(null);
    try {
      await api.act(id, action, body);
    } catch (err) {
      setError(messageOf(err));
      return;
    } finally {
      acting.current = false;
    }
    setPending(null);
    setReason("");
    try {
      await refreshAfterWrite(selection);
      const gen = ++readingGen.current;
      const updated = await api.item(id);
      if (readingGen.current !== gen) return;
      setReading({ ...updated, actions: updated.actions ?? [] });
      setReadingOpacity(1);
    } catch (err) {
      setReadingOpacity(1);
      setError(messageOf(err));
    }
  }

  function onAction(id: string, action: ActionId) {
    setError(null);
    if (action === "decline") {
      openItem(id);
      setReason("");
      setPending("decline");
      return;
    }
    if (action === "snooze") {
      openItem(id);
      setPending("snooze");
      return;
    }
    setPending(null);
    void runAct(id, action);
  }

  async function submitAdd(event: FormEvent) {
    event.preventDefault();
    if (selection.kind !== "tray" || acting.current) return;
    const nextTitle = title.trim();
    if (!nextTitle) return;
    const input: { tray: string; title: string; body?: string } = { tray: selection.ref, title: nextTitle };
    const nextBody = body.trim();
    if (nextBody) input.body = nextBody;
    acting.current = true;
    setError(null);
    try {
      await api.add(input);
    } catch (err) {
      setError(messageOf(err));
      return;
    } finally {
      acting.current = false;
    }
    setTitle("");
    setBody("");
    setAdding(false);
    try {
      await refreshAfterWrite(selection);
    } catch (err) {
      setError(messageOf(err));
    }
  }

  function closeItem() {
    readingGen.current += 1;
    setReading(null);
    setPending(null);
    setError(null);
    setCopied(false);
  }

  useEffect(() => {
    if (!reading?.id) return;
    document.querySelector(`[data-item-id="${CSS.escape(reading.id)}"]`)?.scrollIntoView({ block: "nearest" });
  }, [reading?.id]);

  async function copyRef(text: string) {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch (err) {
      setError(messageOf(err));
    }
  }

  const readingFacts = reading ? [
    fact("Agent", reading.agentSessionId),
    fact("Updated", formatWhen(reading.updatedAt)),
    fact("Due", reading.dueDate),
    fact("Accepted", formatWhen(reading.acceptedAt)),
    fact("Declined", formatWhen(reading.declinedAt)),
    fact("Snoozed until", formatWhen(reading.snoozeUntil)),
    fact("Completed", formatWhen(reading.completedAt ?? reading.doneAt)),
    fact("Archived", formatWhen(reading.archivedAt)),
    fact("Reason", reading.declineReason),
    fact("Note", reading.completionMessage),
  ].filter(Boolean) : [];

  const listBusy = (selection.kind === "attention" && !booted) || trayPhase === "loading";
  const stillLoading = !booted || trayPhase === "loading";

  return (
    <div className={reading ? "shell with-reading" : "shell"} aria-busy={!booted}>
      {stillLoading && <p className="sr-only" role="status">Loading</p>}
      <nav className="sidebar">
        <div className="sidebar-scroll">
        <button type="button" className="nav-row" aria-pressed={selection.kind === "attention"} onClick={chooseAttention}>
          <span>Attention</span>
          {snapshot && snapshot.badge > 0 && <span className="badge">{snapshot.badge}</span>}
        </button>
        {!booted && <SkeletonNav />}
        {booted && (
          <>
            <TrayGroup label="Local" trays={groups.local} selection={selection} onSelect={chooseTray} />
            <TrayGroup label="Directories" trays={groups.directories} selection={selection} onSelect={chooseTray} />
            {session?.signedIn === false && <p className="note remote-note">Sign in with tray login</p>}
            {session?.signedIn === true && snapshot?.remoteStatus === "remote unavailable" && (
              <p className="note remote-note">Remote unavailable</p>
            )}
            <TrayGroup label="Remote" trays={groups.remote} selection={selection} onSelect={chooseTray} />
            <TrayGroup label="Joined" trays={groups.joined} selection={selection} onSelect={chooseTray} />
          </>
        )}
        </div>
        <AppVersion />
      </nav>

      <main className="list" aria-busy={listBusy}>
        {error && !reading && <p className="alert list-alert" role="alert">{error}</p>}
        <div className="list-head">
          <h1>{trayName}</h1>
          <div className="filters" role="group" aria-label="Status">
            {statusFilters.map((filter) => (
              <button
                key={filter}
                type="button"
                className="chip"
                aria-pressed={statusFilter === filter}
                onClick={() => setStatusFilter(filter)}
              >
                {filter === "all" ? "All" : filter[0].toUpperCase() + filter.slice(1)}
              </button>
            ))}
          </div>
        </div>
        {selection.kind === "attention" && !booted && <SkeletonList />}
        {selection.kind === "attention" && booted && (
          <BucketList
            buckets={attention}
            filter={statusFilter}
            empty="Nothing needs attention."
            itemKey={(row) => row.itemId}
            renderItem={(row) => (
              <AttentionRow
                row={row}
                tray={row.tray}
                selected={reading?.id === row.itemId}
                onOpen={() => openItem(row.itemId)}
                onAction={(action) => onAction(row.itemId, action)}
              />
            )}
          />
        )}
        {selection.kind === "tray" && (
          <>
            {trayPhase === "loading" && <SkeletonList />}
            {trayPhase === "ready" && (
            <BucketList
              buckets={buckets}
              filter={statusFilter}
              empty="This tray is empty."
              itemKey={(item) => item.id}
              renderItem={(item) => (
                <ItemRow
                  item={item}
                  selected={reading?.id === item.id}
                  onOpen={() => openItem(item.id)}
                  onAction={(action) => onAction(item.id, action)}
                />
              )}
            />
            )}
            <form className="add" onSubmit={(event) => void submitAdd(event)}>
              <input
                aria-label={`Add to ${trayName}`}
                placeholder={`Add to ${trayName}`}
                value={title}
                onChange={(event) => setTitle(event.target.value)}
                onFocus={() => setAdding(true)}
              />
              {adding && (
                <textarea
                  aria-label="Body"
                  placeholder="Notes (optional)"
                  rows={3}
                  value={body}
                  onChange={(event) => setBody(event.target.value)}
                />
              )}
              {adding && <button className="btn primary" type="submit">Add</button>}
            </form>
          </>
        )}
      </main>

      {reading && (
      <aside className="reading" style={{ opacity: readingOpacity }}>
        <div className="reading-head">
          <span className={`pill ${reading.status}`}>{reading.status}</span>
          <button type="button" className="btn quiet" onClick={closeItem}>Close</button>
        </div>
        {error && <p className="alert" role="alert">{error}</p>}
        <>
            <h2>{reading.title}</h2>
            <p className="meta-line">
              {reading.sourceUser && <span>{reading.sourceUser}</span>}
              {formatWhen(reading.createdAt) && <span>{formatWhen(reading.createdAt)}</span>}
              {reading.clipboard && (
                <button type="button" className="btn link" onClick={() => void copyRef(reading.clipboard ?? "")}>
                  {copied ? "Copied" : "Copy ref"}
                </button>
              )}
            </p>
            {readingFacts.length > 0 && <dl className="facts">{readingFacts}</dl>}
            {reading.body && <p className="body">{reading.body}</p>}
            <div className="reading-actions">
              <ActionBar actions={reading.actions ?? []} onAction={(action) => onAction(reading.id, action)} />
            </div>
            {pending === "decline" && (
              <form
                className="followup"
                onSubmit={(event) => {
                  event.preventDefault();
                  void runAct(reading.id, "decline", { reason });
                }}
              >
                <input
                  aria-label="Reason"
                  placeholder="Reason (optional)"
                  value={reason}
                  onChange={(event) => setReason(event.target.value)}
                />
                <button className="btn" type="submit">Decline</button>
              </form>
            )}
            {pending === "snooze" && (
              <div className="followup">
                {snoozePresets.map(([preset, label]) => (
                  <button
                    key={preset}
                    type="button"
                    className="btn"
                    onClick={() => void runAct(reading.id, "snooze", { until: snoozeUntil(preset, new Date()).toISOString() })}
                  >
                    {label}
                  </button>
                ))}
              </div>
            )}
        </>
      </aside>
      )}
    </div>
  );
}

function SkeletonNav() {
  const groups = [["68%", "46%", "74%"], ["52%", "40%"]];
  return (
    <div aria-hidden="true">
      {groups.map((widths, group) => (
        <div key={group}>
          <span className="skeleton skeleton-label" />
          {widths.map((width) => (
            <div key={width} className="skeleton-nav">
              <span className="skeleton skeleton-line" style={{ width }} />
            </div>
          ))}
        </div>
      ))}
    </div>
  );
}

function SkeletonList() {
  const widths = ["52%", "74%", "40%", "63%"];
  return (
    <div className="item-list" aria-hidden="true">
      {widths.map((width) => (
        <div key={width} className="item skeleton-item">
          <div className="item-line">
            <span className="skeleton skeleton-pill" />
            <span className="skeleton skeleton-line" style={{ width }} />
          </div>
          <span className="skeleton skeleton-meta" />
        </div>
      ))}
    </div>
  );
}

function AppVersion() {
  const version = useAppVersion();
  if (!version) return null;
  return <p className="app-version">{version}</p>;
}

function fact(label: string, value: string | undefined) {
  if (!value) return null;
  return (
    <div key={label}>
      <dt>{label}</dt>
      <dd>{value}</dd>
    </div>
  );
}

function groupTrays(trays: Tray[]) {
  const local: Tray[] = [];
  const directories: Tray[] = [];
  const remote: Tray[] = [];
  const joined: Tray[] = [];
  for (const tray of trays) {
    const kind = tray.kind || (tray.scope === "remote" ? "remote" : "global");
    if (kind === "dir" || kind === "branch") directories.push(tray);
    else if (kind === "joined") joined.push(tray);
    else if (kind === "remote") remote.push(tray);
    else local.push(tray);
  }
  return { local, directories, remote, joined };
}

function bucketItems(items: Item[]): Record<StatusBucket, Item[]> {
  const out: Record<StatusBucket, Item[]> = { pending: [], open: [], completed: [] };
  for (const item of items) out[statusBucket(item.status)].push(item);
  return out;
}

type AttentionItem = SnapshotRow & { tray: string };

function attentionRows(sections: Snapshot["sections"]): Record<StatusBucket, AttentionItem[]> {
  const out: Record<StatusBucket, AttentionItem[]> = { pending: [], open: [], completed: [] };
  for (const section of sections) {
    for (const row of section.rows ?? []) {
      const status = row.scope === "remote" ? "pending" : "open";
      out[statusBucket(status)].push({ ...row, tray: trayHeading(section.header) });
    }
  }
  return out;
}

function BucketList<T>({
  buckets,
  filter,
  empty,
  itemKey,
  renderItem,
}: {
  buckets: Record<StatusBucket, T[]>;
  filter: StatusFilter;
  empty: string;
  itemKey: (item: T) => string;
  renderItem: (item: T) => ReactNode;
}) {
  const order: StatusBucket[] = ["pending", "open", "completed"];
  const labels: Record<StatusBucket, string> = { pending: "Pending", open: "Open", completed: "Completed" };
  const any = order.some((bucket) => buckets[bucket].length > 0);
  if (!any) return <p className="empty-list">{empty}</p>;
  const shown = order.filter((bucket) => (filter === "all" || filter === bucket) && buckets[bucket].length > 0);
  if (shown.length === 0) return <p className="empty-list">Nothing in this filter.</p>;
  return (
    <>
      {shown.map((bucket) => (
        <section key={bucket}>
          <h2 className="section-label">{labels[bucket]} {buckets[bucket].length}</h2>
          <div className="item-list">
            {buckets[bucket].map((item) => (
              <div key={itemKey(item)} className="item-slot">
                {renderItem(item)}
              </div>
            ))}
          </div>
        </section>
      ))}
    </>
  );
}

function TrayGroup({
  label,
  trays,
  selection,
  onSelect,
  note,
}: {
  label: string;
  trays: Tray[];
  selection: Selection;
  onSelect: (ref: string) => void;
  note?: string;
}) {
  if (trays.length === 0 && !note) return null;
  return (
    <>
      <p className="group-label">{label}</p>
      {note && <p className="note">{note}</p>}
      {trays.map((tray) => {
        const nav = tray.kind === "joined" ? joinedNav(tray) : { title: tray.name, detail: tray.kind === "branch" ? tray.detail ?? "" : "" };
        return (
        <button
          key={`${tray.scope}:${tray.ref}`}
          type="button"
          className="nav-row"
          title={tray.kind === "joined" ? tray.name : tray.detail || tray.name}
          aria-pressed={selection.kind === "tray" && selection.ref === tray.ref}
          onClick={() => onSelect(tray.ref)}
        >
          <span className="nav-name">
            <span className="nav-title">{nav.title}</span>
            {nav.detail && <span className="nav-detail">{nav.detail}</span>}
          </span>
          <span className="count">{tray.openCount}</span>
        </button>
        );
      })}
    </>
  );
}

function ItemRow({
  item,
  selected,
  onOpen,
  onAction,
}: {
  item: Item;
  selected: boolean;
  onOpen: () => void;
  onAction: (action: ActionId) => void;
}) {
  return (
    <article
      className="item"
      data-item-id={item.id}
      data-selected={selected ? "true" : "false"}
      onClick={onOpen}
    >
      <div className="item-line">
        <span className={`pill ${item.status}`}>{item.status}</span>
        <span className="item-title">{item.title}</span>
      </div>
      <div className="item-meta">
        {item.sourceUser && <span>{item.sourceUser}</span>}
        {formatWhen(item.createdAt) && <span>{formatWhen(item.createdAt)}</span>}
      </div>
      <ActionBar compact actions={item.actions ?? []} onAction={onAction} />
    </article>
  );
}

function AttentionRow({
  row,
  tray,
  selected,
  onOpen,
  onAction,
}: {
  row: SnapshotRow;
  tray: string;
  selected: boolean;
  onOpen: () => void;
  onAction: (action: ActionId) => void;
}) {
  return (
    <article className="item" data-item-id={row.itemId} data-selected={selected ? "true" : "false"} onClick={onOpen}>
      <div className="item-line">
        <span className={`pill ${row.scope === "remote" ? "pending" : "open"}`}>
          {row.scope === "remote" ? "pending" : "open"}
        </span>
        <span className="item-title">{row.title}</span>
      </div>
      <div className="item-meta">
        <span>{tray}</span>
        {formatWhen(row.createdAt) && <span>{formatWhen(row.createdAt)}</span>}
      </div>
      <ActionBar compact actions={attentionActions(row.scope)} onAction={onAction} />
    </article>
  );
}
