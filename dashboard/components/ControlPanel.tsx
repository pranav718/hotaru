"use client";

import { useState } from "react";
import { ChevronDown } from "lucide-react";

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

  const req = async (endpoint: string, method: string, label: string, reset: () => void) => {
    setLoading(true);
    try {
      const res = await fetch(`http://127.0.0.1:${targetPort}${endpoint}`, { method });
      const text = await res.text();
      logResult(`${label} → ${res.status} ${text}`);
      reset();
    } catch (err: unknown) {
      logResult(`${label} error: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      setLoading(false);
    }
  };

  const handleSet = (e: React.FormEvent) => {
    e.preventDefault();
    if (setKey && setValue) {
      req(`/set?key=${encodeURIComponent(setKey)}&value=${encodeURIComponent(setValue)}`, "POST", `SET ${setKey}=${setValue}`, () => {
        setSetKey("");
        setSetValue("");
      });
    }
  };

  const handleGet = (e: React.FormEvent) => {
    e.preventDefault();
    if (getKey) {
      req(`/get?key=${encodeURIComponent(getKey)}`, "GET", `GET ${getKey}`, () => setGetKey(""));
    }
  };

  const handleDel = (e: React.FormEvent) => {
    e.preventDefault();
    if (delKey) {
      req(`/del?key=${encodeURIComponent(delKey)}`, "DELETE", `DEL ${delKey}`, () => setDelKey(""));
    }
  };

  const handleJoin = (e: React.FormEvent) => {
    e.preventDefault();
    if (joinId && joinAddr) {
      req(`/join?id=${joinId}&addr=${encodeURIComponent(joinAddr)}`, "POST", `JOIN node-${joinId} (${joinAddr})`, () => {
        setJoinId("");
        setJoinAddr("");
      });
    }
  };

  const handleLeave = (e: React.FormEvent) => {
    e.preventDefault();
    if (leaveId) {
      req(`/leave?id=${leaveId}`, "POST", `LEAVE node-${leaveId}`, () => setLeaveId(""));
    }
  };

  return (
    <div className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4 space-y-4 backdrop-blur-md font-mono text-xs">
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-zinc-800/80 pb-3">
        <div className="flex items-center gap-3">
          <div className="text-zinc-200 font-semibold tracking-wide">
            console
          </div>

          <div className="flex bg-zinc-950 p-0.5 rounded-lg border border-zinc-800">
            <button
              onClick={() => setActiveTab("kv")}
              className={`px-2.5 py-1 rounded-md text-[11px] transition-colors ${
                activeTab === "kv" ? "bg-zinc-800 text-zinc-100 font-medium" : "text-zinc-500 hover:text-zinc-300"
              }`}
            >
              kv store
            </button>
            <button
              onClick={() => setActiveTab("membership")}
              className={`px-2.5 py-1 rounded-md text-[11px] transition-colors ${
                activeTab === "membership" ? "bg-zinc-800 text-zinc-100 font-medium" : "text-zinc-500 hover:text-zinc-300"
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
              className="appearance-none bg-zinc-900 border border-zinc-800 rounded-lg pl-3 pr-8 py-1 text-zinc-200 hover:text-zinc-100 hover:border-zinc-700 hover:bg-zinc-800/80 focus:outline-none focus:border-zinc-600 transition-all text-xs font-mono cursor-pointer shadow-sm"
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
          <form onSubmit={handleSet} className="space-y-2 p-3 bg-zinc-950/50 rounded-lg border border-zinc-800/80 flex flex-col justify-between">
            <div className="text-[11px] text-[#C9F27D] font-semibold">
              SET proposal
            </div>
            <div className="space-y-1.5 my-1">
              <input type="text" placeholder="key" value={setKey} onChange={(e) => setSetKey(e.target.value)} className="w-full bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 text-xs" />
              <input type="text" placeholder="value" value={setValue} onChange={(e) => setSetValue(e.target.value)} className="w-full bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 text-xs" />
            </div>
            <button type="submit" disabled={loading || !setKey || !setValue} className="w-full bg-[#C9F27D]/10 border border-[#C9F27D]/30 hover:bg-[#C9F27D]/20 text-[#C9F27D] rounded py-1 transition-colors disabled:opacity-40 text-xs font-medium">
              execute SET
            </button>
          </form>

          <form onSubmit={handleGet} className="space-y-2 p-3 bg-zinc-950/50 rounded-lg border border-zinc-800/80 flex flex-col justify-between">
            <div className="text-[11px] text-zinc-300 font-semibold">
              GET read
            </div>
            <div className="my-1">
              <input type="text" placeholder="key" value={getKey} onChange={(e) => setGetKey(e.target.value)} className="w-full bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 text-xs" />
            </div>
            <button type="submit" disabled={loading || !getKey} className="w-full bg-zinc-800/80 border border-zinc-700/60 hover:bg-zinc-700/80 text-zinc-100 rounded py-1 transition-colors disabled:opacity-40 text-xs font-medium">
              execute GET
            </button>
          </form>

          <form onSubmit={handleDel} className="space-y-2 p-3 bg-zinc-950/50 rounded-lg border border-zinc-800/80 flex flex-col justify-between">
            <div className="text-[11px] text-red-400 font-semibold">
              DEL delete
            </div>
            <div className="my-1">
              <input type="text" placeholder="key" value={delKey} onChange={(e) => setDelKey(e.target.value)} className="w-full bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 text-xs" />
            </div>
            <button type="submit" disabled={loading || !delKey} className="w-full bg-red-500/10 border border-red-500/30 hover:bg-red-500/20 text-red-400 rounded py-1 transition-colors disabled:opacity-40 text-xs font-medium">
              execute DEL
            </button>
          </form>
        </div>
      )}

      {activeTab === "membership" && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <form onSubmit={handleJoin} className="space-y-2 p-3 bg-zinc-950/50 rounded-lg border border-zinc-800/80">
            <div className="text-[11px] text-[#C9F27D] font-semibold">
              JOIN node
            </div>
            <div className="grid grid-cols-3 gap-2">
              <input type="number" placeholder="id" value={joinId} onChange={(e) => setJoinId(e.target.value)} className="col-span-1 bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 text-xs min-w-0" />
              <input type="text" placeholder="addr (127.0.0.1:8003)" value={joinAddr} onChange={(e) => setJoinAddr(e.target.value)} className="col-span-2 bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 text-xs min-w-0" />
            </div>
            <button type="submit" disabled={loading || !joinId || !joinAddr} className="w-full bg-[#C9F27D]/10 border border-[#C9F27D]/30 hover:bg-[#C9F27D]/20 text-[#C9F27D] rounded py-1 transition-colors disabled:opacity-40 text-xs font-medium">
              propose JOIN
            </button>
          </form>

          <form onSubmit={handleLeave} className="space-y-2 p-3 bg-zinc-950/50 rounded-lg border border-zinc-800/80">
            <div className="text-[11px] text-red-400 font-semibold">
              LEAVE node
            </div>
            <input type="number" placeholder="node id (e.g. 3)" value={leaveId} onChange={(e) => setLeaveId(e.target.value)} className="w-full bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 text-xs min-w-0" />
            <button type="submit" disabled={loading || !leaveId} className="w-full bg-red-500/10 border border-red-500/30 hover:bg-red-500/20 text-red-400 rounded py-1 transition-colors disabled:opacity-40 text-xs font-medium">
              propose LEAVE
            </button>
          </form>
        </div>
      )}

      <div data-lenis-prevent className="bg-zinc-950 rounded-lg p-3 border border-zinc-800/80 h-[140px] overflow-y-auto space-y-1.5 text-[11px] text-zinc-300 font-mono scrollbar-thin">
        {outputLog.length === 0 ? (
          <div className="text-zinc-600">console output ready...</div>
        ) : (
          outputLog.map((log, i) => (
            <div key={i} className="leading-relaxed break-all">
              {log}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
