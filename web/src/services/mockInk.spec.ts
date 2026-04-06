import { generateReplyWithMockService } from "@/services/mockInk";

describe("mock ink services", () => {
  it("generates a stable reply and supports explicit error simulation", async () => {
    await expect(
      generateReplyWithMockService({
        prompt: "请帮我整理一句提醒",
      }),
    ).resolves.toContain("当然可以。你可以这样写：");

    await expect(
      generateReplyWithMockService({
        prompt: "[error] trigger failure",
      }),
    ).rejects.toThrow("暂时没能生成回复，请稍后重试。");
  });
});
