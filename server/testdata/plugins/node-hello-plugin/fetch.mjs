import process from "node:process";

async function readJSON() {
  const chunks = [];
  for await (const chunk of process.stdin) {
    chunks.push(chunk);
  }

  const raw = Buffer.concat(chunks).toString("utf8").trim();
  return raw ? JSON.parse(raw) : {};
}

const payload = await readJSON();
const sourceName = String(payload.workspaceConfig?.sourceName ?? "Node Hello Source").trim();
const message = String(
  payload.scheduleConfig?.message ?? "Hello from the Node fixture plugin.",
).trim();
const tone = String(payload.scheduleConfig?.tone ?? "plain").trim();
const repeatValue = Number.parseInt(String(payload.scheduleConfig?.repeat ?? "1"), 10);
const repeat = Number.isFinite(repeatValue) && repeatValue > 0 ? repeatValue : 1;
const includeTriggeredAt = Boolean(payload.workspaceConfig?.includeTriggeredAt);

const repeated = Array.from({ length: repeat }, () => message);
if (tone === "verbose" && includeTriggeredAt && payload.trigger?.triggeredAt) {
  repeated.push(`Triggered at: ${payload.trigger.triggeredAt}`);
}

process.stdout.write(
  JSON.stringify({
    title: `${sourceName} Digest`,
    content: repeated.join("\n"),
    sourceLabel: sourceName,
  }),
);
