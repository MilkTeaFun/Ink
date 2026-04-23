interface ReplyPayload {
  prompt: string;
}

import { translate } from "@/i18n";

const NETWORK_DELAY_MS = 80;

function wait(ms = NETWORK_DELAY_MS) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

export async function generateReplyWithMockService({ prompt }: ReplyPayload): Promise<string> {
  await wait();

  if (prompt.toLowerCase().includes("[error]")) {
    throw new Error(translate("mockInk.error"));
  }

  const trimmedPrompt = prompt.trim().replace(/\s+/g, " ");
  const summary = translate("mockInk.summary");
  const promptSuffix = trimmedPrompt.endsWith("。") || trimmedPrompt.endsWith(".") ? "" : "。";

  return translate("mockInk.reply", {
    prompt: `${trimmedPrompt}${promptSuffix}`,
    summary,
  });
}
