import { afterEach, beforeEach, expect, it, vi } from "vitest";

import { submitFeedbackToAdmin } from "@/services/feedback";

const fetchMock = vi.fn<typeof fetch>();

describe("feedback service", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    fetchMock.mockReset();
  });

  it("submits feedback through the authenticated feedback endpoint", async () => {
    fetchMock.mockResolvedValueOnce(new Response(null, { status: 204 }));

    await expect(submitFeedbackToAdmin("access-token", "希望加一个反馈入口")).resolves.toBe(
      undefined,
    );

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/feedback/print",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ content: "希望加一个反馈入口" }),
      }),
    );
  });
});
