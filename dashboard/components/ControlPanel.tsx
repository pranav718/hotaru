"use client";

import { useState } from "react";
import { Terminal, Send, Search, Trash2, UserPlus, UserMinus, ChevronDown } from "lucide-react";

export function ControlPanel() {
  const [targetPort, setTargetPort] = useState(8010);
  const [activeTab, setActiveTab] = useState<"kv" | "membership">("kv");

  const [setKey, setSetKey] = useState("");
  const [setValue, setSetValue] = useState("");
  const [getKey, setGetKey] = useState("");
  const [delKey, setDelKey] = useState("");

  const [joinId, setJoinId] = useState("");
  const [joinAddr, setJoinAddr] = useState("");
  const [leaveId, setLeaveId] = useState("");

  const [outputLog, setOutputLog] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);

  const logResult = (msg: string) => {
    const ts = new Date().toLocaleTimeString();
    setOutputLog((prev) => [`[${ts}] ${msg}`, ...prev.slice(0, 49)]);
  };

  const handleSet = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!setKey || !setValue) return;
    setLoading(true);
    try {
      const res = await fetch(
        `http://127.0.0.1:${targetPort}/set?key=${encodeURIComponent(
          setKey
        )}&value=${encodeURIComponent(setValue)}`,
        { method: "POST" }
      );
      const text = await res.text();
      logResult(`SET ${setKey}=${setValue} → ${res.status} ${text}`);
      setSetKey("");
      setSetValue("");
    } catch (err: unknown) {
      logResult(`SET error: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      setLoading(false);
    }
  };

  const handleGet = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!getKey) return;
    setLoading(true);
    try {
      const res = await fetch(
        `http://127.0.0.1:${targetPort}/get?key=${encodeURIComponent(getKey)}`
      );
      const text = await res.text();
      logResult(`GET ${getKey} → ${res.status} "${text}"`);
      setGetKey("");
    } catch (err: unknown) {
      logResult(`GET error: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      setLoading(false);
    }
  };

  const handleDel = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!delKey) return;
    setLoading(true);
    try {
      const res = await fetch(
        `http://127.0.0.1:${targetPort}/del?key=${encodeURIComponent(delKey)}`,
        { method: "DELETE" }
      );
      const text = await res.text();
      logResult(`DEL ${delKey} → ${res.status} ${text}`);
      setDelKey("");
    } catch (err: unknown) {
      logResult(`DEL error: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      setLoading(false);
    }
  };

  const handleJoin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!joinId || !joinAddr) return;
    setLoading(true);
    try {
      const res = await fetch(
        `http://127.0.0.1:${targetPort}/join?id=${joinId}&addr=${encodeURIComponent(
          joinAddr
        )}`,
        { method: "POST" }
      );
      const text = await res.text();
      logResult(`JOIN node-${joinId} (${joinAddr}) → ${res.status} ${text}`);
      setJoinId("");
      setJoinAddr("");
    } catch (err: unknown) {
      logResult(`JOIN error: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      setLoading(false);
    }
  };

  const handleLeave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!leaveId) return;
    setLoading(true);
    try {
      const res = await fetch(
        `http://127.0.0.1:${targetPort}/leave?id=${leaveId}`,
        { method: "POST" }
      );
      const text = await res.text();
      logResult(`LEAVE node-${leaveId} → ${res.status} ${text}`);
      setLeaveId("");
    } catch (err: unknown) {
      logResult(`LEAVE error: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4 space-y-4 backdrop-blur-md font-mono text-xs">
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-zinc-800/80 pb-3">
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1.5 text-cyan-400 font-semibold">
            <Terminal className="h-4 w-4" />
            <span className="uppercase tracking-wide">console</span>
          </div>

          <div className="flex bg-zinc-950 p-0.5 rounded-lg border border-zinc-800">
            <button
              onClick={() => setActiveTab("kv")}
              className={`px-2.5 py-1 rounded-md text-[11px] transition-colors ${
                activeTab === "kv"
                  ? "bg-zinc-800 text-zinc-100 font-medium"
                  : "text-zinc-500 hover:text-zinc-300"
              }`}
            >
              kv store
            </button>
            <button
              onClick={() => setActiveTab("membership")}
              className={`px-2.5 py-1 rounded-md text-[11px] transition-colors ${
                activeTab === "membership"
                  ? "bg-zinc-800 text-zinc-100 font-medium"
                  : "text-zinc-500 hover:text-zinc-300"
              }`}
            >
              membership
            </button>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <span className="text-[11px] text-zinc-500 font-mono">target:</span>
          <div className="relative inline-block">
            <select
              value={targetPort}
              onChange={(e) => setTargetPort(Number(e.target.value))}
              className="appearance-none bg-zinc-900 border border-zinc-800 rounded-lg pl-3 pr-8 py-1 text-zinc-200 hover:text-zinc-100 hover:border-zinc-700 hover:bg-zinc-800/80 focus:outline-none focus:border-cyan-500/50 transition-all text-xs font-mono cursor-pointer shadow-sm"
            >
              <option value={8010} className="bg-zinc-900 text-zinc-200">
                node-0 (:8010)
              </option>
              <option value={8011} className="bg-zinc-900 text-zinc-200">
                node-1 (:8011)
              </option>
              <option value={8012} className="bg-zinc-900 text-zinc-200">
                node-2 (:8012)
              </option>
            </select>
            <ChevronDown className="absolute right-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-zinc-400 pointer-events-none" />
          </div>
        </div>
      </div>

      {activeTab === "kv" && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          <form
            onSubmit={handleSet}
            className="space-y-2 p-3 bg-zinc-950/50 rounded-lg border border-zinc-800/80 flex flex-col justify-between"
          >
            <div className="text-[11px] text-[#C9F27D] font-semibold flex items-center gap-1.5">
              <Send className="h-3 w-3" /> SET proposal
            </div>
            <div className="space-y-1.5 my-1">
              <input
                type="text"
                placeholder="key"
                value={setKey}
                onChange={(e) => setSetKey(e.target.value)}
                className="w-full bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 text-xs"
              />
              <input
                type="text"
                placeholder="value"
                value={setValue}
                onChange={(e) => setSetValue(e.target.value)}
                className="w-full bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 text-xs"
              />
            </div>
            <button
              type="submit"
              disabled={loading || !setKey || !setValue}
              className="w-full bg-[#C9F27D]/10 border border-[#C9F27D]/30 hover:bg-[#C9F27D]/20 text-[#C9F27D] rounded py-1 transition-colors disabled:opacity-40 text-xs"
            >
              execute SET
            </button>
          </form>

          <form
            onSubmit={handleGet}
            className="space-y-2 p-3 bg-zinc-950/50 rounded-lg border border-zinc-800/80 flex flex-col justify-between"
          >
            <div className="text-[11px] text-cyan-400 font-semibold flex items-center gap-1.5">
              <Search className="h-3 w-3" /> GET read
            </div>
            <div className="my-1">
              <input
                type="text"
                placeholder="key"
                value={getKey}
                onChange={(e) => setGetKey(e.target.value)}
                className="w-full bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 text-xs"
              />
            </div>
            <button
              type="submit"
              disabled={loading || !getKey}
              className="w-full bg-cyan-600/20 border border-cyan-500/30 hover:bg-cyan-600/30 text-cyan-300 rounded py-1 transition-colors disabled:opacity-40 text-xs"
            >
              execute GET
            </button>
          </form>

          <form
            onSubmit={handleDel}
            className="space-y-2 p-3 bg-zinc-950/50 rounded-lg border border-zinc-800/80 flex flex-col justify-between"
          >
            <div className="text-[11px] text-rose-400 font-semibold flex items-center gap-1.5">
              <Trash2 className="h-3 w-3" /> DEL delete
            </div>
            <div className="my-1">
              <input
                type="text"
                placeholder="key"
                value={delKey}
                onChange={(e) => setDelKey(e.target.value)}
                className="w-full bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 text-xs"
              />
            </div>
            <button
              type="submit"
              disabled={loading || !delKey}
              className="w-full bg-rose-600/20 border border-rose-500/30 hover:bg-rose-600/30 text-rose-300 rounded py-1 transition-colors disabled:opacity-40 text-xs"
            >
              execute DEL
            </button>
          </form>
        </div>
      )}

      {activeTab === "membership" && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <form
            onSubmit={handleJoin}
            className="space-y-2 p-3 bg-zinc-950/50 rounded-lg border border-zinc-800/80"
          >
            <div className="text-[11px] text-teal-400 font-semibold flex items-center gap-1.5">
              <UserPlus className="h-3 w-3" /> JOIN node
            </div>
            <div className="grid grid-cols-3 gap-2">
              <input
                type="number"
                placeholder="id"
                value={joinId}
                onChange={(e) => setJoinId(e.target.value)}
                className="col-span-1 bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 text-xs min-w-0"
              />
              <input
                type="text"
                placeholder="addr (127.0.0.1:8003)"
                value={joinAddr}
                onChange={(e) => setJoinAddr(e.target.value)}
                className="col-span-2 bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 text-xs min-w-0"
              />
            </div>
            <button
              type="submit"
              disabled={loading || !joinId || !joinAddr}
              className="w-full bg-teal-600/20 border border-teal-500/30 hover:bg-teal-600/30 text-teal-300 rounded py-1 transition-colors disabled:opacity-40 text-xs"
            >
              propose JOIN
            </button>
          </form>

          <form
            onSubmit={handleLeave}
            className="space-y-2 p-3 bg-zinc-950/50 rounded-lg border border-zinc-800/80"
          >
            <div className="text-[11px] text-rose-400 font-semibold flex items-center gap-1.5">
              <UserMinus className="h-3 w-3" /> LEAVE node
            </div>
            <input
              type="number"
              placeholder="node id (e.g. 3)"
              value={leaveId}
              onChange={(e) => setLeaveId(e.target.value)}
              className="w-full bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 text-xs min-w-0"
            />
            <button
              type="submit"
              disabled={loading || !leaveId}
              className="w-full bg-rose-600/20 border border-rose-500/30 hover:bg-rose-600/30 text-rose-300 rounded py-1 transition-colors disabled:opacity-40 text-xs"
            >
              propose LEAVE
            </button>
          </form>
        </div>
      )}

      <div className="bg-zinc-950 rounded-lg p-3 border border-zinc-800/80 max-h-[120px] overflow-y-auto space-y-1 text-[11px] text-zinc-400 scrollbar-thin">
        {outputLog.length === 0 ? (
          <div className="text-zinc-600 italic">console output ready...</div>
        ) : (
          outputLog.map((log, i) => (
            <div key={i} className="leading-tight truncate">
              {log}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
