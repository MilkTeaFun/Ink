import { generateReplyWithMockService, loginWithMockService } from "@/services/mockInk";

describe("mock ink services", () => {
  it("returns a mock user for valid credentials", async () => {
    await expect(
      loginWithMockService({
        email: "demo@example.com",
        password: "safe-password",
      }),
    ).resolves.toMatchObject({
      email: "demo@example.com",
      name: "demo",
    });
  });

  it("rejects invalid credentials", async () => {
    await expect(
      loginWithMockService({
        email: "fail@example.com",
        password: "safe-password",
      }),
    ).rejects.toThrow("邮箱或密码不正确。");
  });

  it("generates different styles and supports explicit error simulation", async () => {
    await expect(
      generateReplyWithMockService({
        prompt: "请帮我整理一句提醒",
        answerStyle: "warm-encouraging",
        noteStyle: "gentle",
        responseLength: "long",
      }),
    ).resolves.toContain("当然，我们把这件事说得再柔和一点。");

    await expect(
      generateReplyWithMockService({
        prompt: "[error] trigger failure",
        answerStyle: "clear-gentle",
        noteStyle: "clean",
        responseLength: "short",
      }),
    ).rejects.toThrow("暂时没能生成回复，请稍后重试。");
  });
});
