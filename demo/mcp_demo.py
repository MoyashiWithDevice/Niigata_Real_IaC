#!/usr/bin/env python3
"""MCP stdio demo driver for IACForge.

Speaks JSON-RPC 2.0 over the MCP stdio transport sequentially (one request
at a time) so that the session state persists between tool calls.
"""
import json
import os
import subprocess
import sys

BIN = sys.argv[1] if len(sys.argv) > 1 else "./iacforge"
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

proc = subprocess.Popen(
    [BIN, "mcp", "--stdio"],
    stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL,
    text=True, bufsize=1, cwd=ROOT,
)

_id = 0


def call(method, params=None, notify=False):
    global _id
    msg = {"jsonrpc": "2.0", "method": method}
    if not notify:
        _id += 1
        msg["id"] = _id
    if params is not None:
        msg["params"] = params
    proc.stdin.write(json.dumps(msg) + "\n")
    proc.stdin.flush()
    if notify:
        return None
    while True:
        line = proc.stdout.readline()
        if not line:
            raise RuntimeError("MCP server closed unexpectedly")
        resp = json.loads(line)
        if resp.get("id") == msg["id"]:
            return resp


def tool(name, args=None):
    r = call("tools/call", {"name": name, "arguments": args or {}})
    content = r.get("result", {}).get("content", [])
    return content[0]["text"] if content else ""


def banner(title):
    print(f"\n>>> {title}")


# --- handshake -------------------------------------------------------------
call("initialize", {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {"name": "iacforge-demo", "version": "0.1.0"},
})
call("notifications/initialized", notify=True)
print("MCP session initialized")

# --- tools/list ------------------------------------------------------------
r = call("tools/list")
tools = sorted(t["name"] for t in r["result"]["tools"])
banner(f"tools/list: {len(tools)} tools available")
for i in range(0, len(tools), 6):
    print("  " + ", ".join(tools[i:i + 6]))

# --- load a model ----------------------------------------------------------
banner('load_yaml demo/core/model.yaml')
print(tool("load_yaml", {"path": "demo/core/model.yaml"}))

banner("graph_summary")
print(tool("graph_summary"))

# --- mutate the graph ------------------------------------------------------
banner("add_entity vm-cache-01 (owner: srv-proxmox-02)")
print(tool("add_entity", {
    "id": "vm-cache-01", "kind": "vm", "name": "Cache VM 01",
    "owner": "srv-proxmox-02", "status": "active",
}))

banner("add_entity app-redis (owner: vm-cache-01, properties via JSON)")
print(tool("add_entity", {
    "id": "app-redis", "kind": "application", "name": "Redis",
    "owner": "vm-cache-01", "status": "active",
    "properties_json": '{"port": 6379, "protocol": "tcp", "version": "7.2"}',
}))

banner("add_relation app-api depends_on app-redis")
print(tool("add_relation", {
    "id": "rel-api-dep-redis", "type": "depends_on",
    "source": "app-api", "target": "app-redis",
}))

banner("query_entities kind=vm")
out = tool("query_entities", {"kind": "vm"})
print("\n".join(out.splitlines()[:20]))
n = len(out.splitlines())
if n > 20:
    print(f"... ({n - 20} more lines)")

banner("who_references app-api")
print(tool("who_references", {"id": "app-api"}))

# --- validate & render -----------------------------------------------------
banner("validate_graph")
d = json.loads(tool("validate_graph"))
print(f"passed={d['passed']} "
      f"(rules: {d['summary']['total_rules']}, findings: {d['summary']['total_findings']}, "
      f"errors: {d['summary']['errors']})")

banner("render_graph format=mermaid kinds=[vm,application]")
print(tool("render_graph", {"format": "mermaid", "kinds": ["vm", "application"]}))

proc.stdin.close()
proc.terminate()
