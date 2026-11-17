import { useEffect, useRef, useState, type FormEvent } from "react";
import { ActionBar } from "./ActionBar";
import type { ActionId } from "./actions";
import {
  connect,
  type Item,
  type Session,
  type Snapshot,
  type Tray,
} from "./api";
import { snoozeUntil } from "./snooze";

type Selection = { kind: "attention" } | { kind: "tray"; ref: string };
type Filter = "open" | "all";

const POLL_MS = 30_000;

const snoozePresets = [
  ["hour", "Hour"],
  ["tomorrow", "Tomorrow"],
  ["week", "Next week"],
] as const;

function messageOf(err: unknown): string {
  return err instanceof Error && err.message ? err.message : "request failed";
}

function listKey(selection: Selection, filter: Filter): string {
  return selection.kind === "tray" ? `${selection.ref}\n${filter}` : "";
}

export default function App() {
  const apiConfig = window.__TRAY_API__;
  if (!apiConfig) return <p>tray is not installed</p>;
  return <TrayWindow base={apiConfig.base} token={apiConfig.token} />;
}

function TrayWindow({ base, token }: { base: string; token: string }) {
  const [api] = useState(() => connect(base, token));
  const [session, setSession] = useState<Session | null>(null);
  const [snapshot, setSnapshot] = useState<Snapshot | null>(null);
  const [trays, setTrays] = useState<Tray[]>([]);
  const [selection, setSelection] = useState<Selection>({ kind: "attention" });
  const [filter, setFilter] = useState<Filter>("open");
  const [loaded, setLoaded] = useState<{ key: string; items: Item[] }>({
    key: "",
    items: [],
  });
  const [reading, setReading] = useState<Item | null>(null);
  const [readingOpacity, setReadingOpacity] = useState(1);
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState<"decline" | "snooze" | null>(null);
  const [reason, setReason] = useState("");
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const listGen = useRef(0);
  const readingGen = useRef(0);
  const acting = useRef(false);

  useEffect(() => {
    let cancelled = false;
    const fail = (err: unknown) => {
      if (!cancelled) setError(messageOf(err));
    };
    api
      .session()
      .then((next) => {
        if (!cancelled) setSession(next);
      })
      .catch(fail);
    api
      .snapshot()
      .then((next) => {
        if (!cancelled) setSnapshot(next);
      })
      .catch(fail);
    api
      .trays()
      .then((next) => {
        if (!cancelled) setTrays(Array.isArray(next) ? next : []);
      })
      .catch(fail);

    const timer = setInterval(() => {
      api
        .snapshot()
        .then((next) => {
          if (!cancelled) setSnapshot(next);
        })
        .catch(fail);
      api
        .session()
        .then((next) => {
          if (!cancelled) setSession(next);
        })
        .catch(() => {});
      api
        .trays()
        .then((next) => {
          if (!cancelled && Array.isArray(next)) setTrays(next);
        })
        .catch(() => {});
    }, POLL_MS);

    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [api]);

  useEffect(() => {
    readingGen.current += 1;
    setReading(null);
    setReadingOpacity(1);
    setPending(null);
    setReason("");
    setTitle("");
    setBody("");
  }, [selection]);

  useEffect(() => {
    if (selection.kind !== "tray") {
      listGen.current += 1;
      return;
    }
    let cancelled = false;
    const ref = selection.ref;
    const current = filter;
    const key = listKey(selection, current);

    const load = () => {
      const gen = ++listGen.current;
      return api.items(ref, current).then((next) => {
        if (cancelled || listGen.current !== gen) return;
        setLoaded({ key, items: Array.isArray(next) ? next : [] });
      });
    };

    load().catch((err: unknown) => {
      if (!cancelled) setError(messageOf(err));
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
  }, [api, selection, filter]);

  const visibleItems =
    loaded.key === listKey(selection, filter) ? loaded.items : [];
  const selectedTray =
    selection.kind === "tray"
      ? trays.find((tray) => tray.ref === selection.ref)
      : undefined;
  const trayName =
    selectedTray?.name ?? (selection.kind === "tray" ? selection.ref : "");
  const localTrays = trays.filter((tray) => tray.scope === "local");
  const remoteTrays = trays.filter((tray) => tray.scope === "remote");
  const sections =
    snapshot && Array.isArray(snapshot.sections) ? snapshot.sections : [];

  function chooseAttention() {
    if (selection.kind === "attention") return;
    setFilter("open");
    setSelection({ kind: "attention" });
  }

  function chooseTray(ref: string) {
    if (selection.kind === "tray" && selection.ref === ref) return;
    setFilter("open");
    setSelection({ kind: "tray", ref });
  }

  function openItem(id: string) {
    const gen = ++readingGen.current;
    setReadingOpacity(0);
    setError(null);
    setPending(null);
    api
      .item(id)
      .then((item) => {
        if (readingGen.current !== gen) return;
        setReading({ ...item, actions: item.actions ?? [] });
        setReadingOpacity(1);
      })
      .catch((err: unknown) => {
        if (readingGen.current !== gen) return;
        setReadingOpacity(1);
        setError(messageOf(err));
      });
  }

  async function refreshAfterWrite(current: Selection, currentFilter: Filter) {
    const gen = ++listGen.current;
    const key = listKey(current, currentFilter);
    const [snapR, traysR, itemsR] = await Promise.allSettled([
      api.snapshot(),
      api.trays(),
      current.kind === "tray"
        ? api.items(current.ref, currentFilter)
        : Promise.resolve(null),
    ]);
    if (snapR.status === "fulfilled") setSnapshot(snapR.value);
    if (traysR.status === "fulfilled") {
      setTrays(Array.isArray(traysR.value) ? traysR.value : []);
    }
    if (
      current.kind === "tray" &&
      itemsR.status === "fulfilled" &&
      listGen.current === gen
    ) {
      setLoaded({
        key,
        items: Array.isArray(itemsR.value) ? itemsR.value : [],
      });
    }
    if (snapR.status === "rejected") throw snapR.reason;
    if (current.kind === "tray" && itemsR.status === "rejected") {
      throw itemsR.reason;
    }
  }

  async function runAct(action: ActionId, body?: unknown) {
    if (!reading || acting.current) return;
    const id = reading.id;
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
      await refreshAfterWrite(selection, filter);
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

  function onAction(action: ActionId) {
    setError(null);
    if (action === "decline") {
      setReason("");
      setPending("decline");
      return;
    }
    if (action === "snooze") {
      setPending("snooze");
      return;
    }
    setPending(null);
    void runAct(action);
  }

  async function submitAdd(event: FormEvent) {
    event.preventDefault();
    if (selection.kind !== "tray" || acting.current) return;
    const nextTitle = title.trim();
    if (!nextTitle) return;
    const input: { tray: string; title: string; body?: string } = {
      tray: selection.ref,
      title: nextTitle,
    };
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
    try {
      await refreshAfterWrite(selection, filter);
    } catch (err) {
      setError(messageOf(err));
    }
  }

  return (
    <div className="shell">
      <nav className="pane">
        <button
          type="button"
          aria-pressed={selection.kind === "attention"}
          onClick={chooseAttention}
        >
          Attention{snapshot ? ` ${snapshot.badge}` : ""}
        </button>
        {localTrays.length > 0 && <p className="muted">Local</p>}
        {localTrays.map((tray) => (
          <TrayButton
            key={`local:${tray.ref}`}
            tray={tray}
            pressed={selection.kind === "tray" && selection.ref === tray.ref}
            onSelect={chooseTray}
          />
        ))}
        {session?.signedIn === false && (
          <p className="muted">Sign in with tray login</p>
        )}
        {session?.signedIn === true &&
          snapshot?.remoteStatus === "remote unavailable" && (
            <p className="muted">remote unavailable</p>
          )}
        {session?.signedIn === true &&
          remoteTrays.map((tray) => (
            <TrayButton
              key={`remote:${tray.ref}`}
              tray={tray}
              pressed={selection.kind === "tray" && selection.ref === tray.ref}
              onSelect={chooseTray}
            />
          ))}
      </nav>

      <main className="pane">
        {selection.kind === "attention" &&
          sections.map((section, index) => (
            <section key={`${section.header}:${index}`}>
              <h2>{section.header}</h2>
              {(Array.isArray(section.rows) ? section.rows : []).map((row) => (
                <button
                  key={row.itemId}
                  type="button"
                  aria-pressed={reading?.id === row.itemId}
                  onClick={() => openItem(row.itemId)}
                >
                  {row.title}
                </button>
              ))}
            </section>
          ))}
        {selection.kind === "tray" && (
          <>
            <form onSubmit={(event) => void submitAdd(event)}>
              <input
                aria-label={`Add to ${trayName}`}
                placeholder={`Add to ${trayName}`}
                value={title}
                onChange={(event) => setTitle(event.target.value)}
              />
              <textarea
                aria-label="Body"
                placeholder="Body (optional)"
                rows={3}
                value={body}
                onChange={(event) => setBody(event.target.value)}
              />
              <button type="submit">Add</button>
            </form>
            <button
              type="button"
              aria-pressed={filter === "all"}
              onClick={() =>
                setFilter((current) => (current === "open" ? "all" : "open"))
              }
            >
              {filter === "open" ? "Open" : "All"}
            </button>
            {visibleItems.map((item) => (
              <button
                key={item.id}
                type="button"
                aria-pressed={reading?.id === item.id}
                onClick={() => openItem(item.id)}
              >
                {item.title} <span className="muted">{item.status}</span>
              </button>
            ))}
          </>
        )}
      </main>

      <aside className="reading pane" style={{ opacity: readingOpacity }}>
        {error && <p role="alert">{error}</p>}
        {reading && (
          <>
            <h2>{reading.title}</h2>
            <p>{reading.status}</p>
            {reading.snoozeUntil && (
              <p className="muted">{reading.snoozeUntil}</p>
            )}
            {reading.body && <p className="body">{reading.body}</p>}
            <ActionBar actions={reading.actions ?? []} onAction={onAction} />
            {pending === "decline" && (
              <form
                onSubmit={(event) => {
                  event.preventDefault();
                  void runAct("decline", { reason });
                }}
              >
                <input
                  aria-label="Reason"
                  placeholder="Reason (optional)"
                  value={reason}
                  onChange={(event) => setReason(event.target.value)}
                />
                <button type="submit">Decline</button>
              </form>
            )}
            {pending === "snooze" &&
              snoozePresets.map(([preset, label]) => (
                <button
                  key={preset}
                  type="button"
                  onClick={() =>
                    void runAct("snooze", {
                      until: snoozeUntil(preset, new Date()).toISOString(),
                    })
                  }
                >
                  {label}
                </button>
              ))}
          </>
        )}
      </aside>
    </div>
  );
}

function TrayButton({
  tray,
  pressed,
  onSelect,
}: {
  tray: Tray;
  pressed: boolean;
  onSelect: (ref: string) => void;
}) {
  return (
    <button
      type="button"
      aria-pressed={pressed}
      onClick={() => onSelect(tray.ref)}
    >
      {tray.name} <span className="muted">{tray.openCount}</span>
    </button>
  );
}
