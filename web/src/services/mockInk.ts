import type { AnswerStyle, NoteStyle, ResponseLength, User } from "@/types/workspace";

interface LoginPayload {
  email: string;
  password: string;
}

interface ReplyPayload {
  prompt: string;
  answerStyle: AnswerStyle;
  noteStyle: NoteStyle;
  responseLength: ResponseLength;
}

const NETWORK_DELAY_MS = 80;

function wait(ms = NETWORK_DELAY_MS) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

function buildGreeting(style: AnswerStyle) {
  if (style === "warm-encouraging") {
    return "当然，我们把这件事说得再柔和一点。";
  }

  if (style === "concise-direct") {
    return "可以，下面是更利落的版本。";
  }

  return "当然可以。你可以这样写：";
}

function buildClosing(noteStyle: NoteStyle, responseLength: ResponseLength) {
  const closings = {
    clean: "留一行空白，会更适合打印在纸条上。",
    gentle: "这样读起来会更轻一点，也更像一句贴心提醒。",
    list: "如果你愿意，我也可以帮你改成清单格式。",
  } as const;

  if (responseLength === "short") {
    return "";
  }

  if (responseLength === "medium") {
    return closings[noteStyle];
  }

  return `${closings[noteStyle]} 如果要直接拿去打印，建议保留一句重点和一句收尾。`;
}

export async function loginWithMockService({ email, password }: LoginPayload): Promise<User> {
  await wait();

  if (password === "wrong" || email.includes("fail")) {
    throw new Error("邮箱或密码不正确。");
  }

  return {
    id: "user-ink-demo",
    email,
    name: email.split("@")[0] || "Ink 用户",
  };
}

export async function generateReplyWithMockService({
  prompt,
  answerStyle,
  noteStyle,
  responseLength,
}: ReplyPayload): Promise<string> {
  await wait();

  if (prompt.toLowerCase().includes("[error]")) {
    throw new Error("暂时没能生成回复，请稍后重试。");
  }

  const greeting = buildGreeting(answerStyle);
  const closing = buildClosing(noteStyle, responseLength);
  const trimmedPrompt = prompt.trim().replace(/\s+/g, " ");

  const summary =
    responseLength === "short"
      ? `先把最重要的一件事做好，再记得照顾一下自己。`
      : responseLength === "medium"
        ? `先把最重要的一件事做好，处理完就停一下，晚一点给自己留一点缓冲和休息。`
        : `先把最重要的一件事做好，不需要一次解决所有问题。处理完最关键的部分后，给自己留一点喘息时间，慢一点也没有关系。`;

  return `${greeting}${trimmedPrompt}${trimmedPrompt.endsWith("。") ? "" : "。"}${summary}${closing}`;
}
