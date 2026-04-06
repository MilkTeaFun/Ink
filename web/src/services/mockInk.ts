interface ReplyPayload {
  prompt: string;
}

const NETWORK_DELAY_MS = 80;

function wait(ms = NETWORK_DELAY_MS) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

export async function generateReplyWithMockService({ prompt }: ReplyPayload): Promise<string> {
  await wait();

  if (prompt.toLowerCase().includes("[error]")) {
    throw new Error("暂时没能生成回复，请稍后重试。");
  }

  const trimmedPrompt = prompt.trim().replace(/\s+/g, " ");
  const summary = "先把最重要的一件事做好，处理完就停一下，给自己留一点缓冲和休息。";

  return `当然可以。你可以这样写：${trimmedPrompt}${trimmedPrompt.endsWith("。") ? "" : "。"}${summary}留一行空白，会更适合打印在纸条上。`;
}
