import { afterEach, beforeEach, expect, it, vi } from "vitest";

import {
  createUserWithApi,
  fetchWorkspaceStateWithApi,
  saveWorkspaceStateWithApi,
} from "@/services/workspace";

const fetchMock = vi.fn<typeof fetch>();

describe("workspace service", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    fetchMock.mockReset();
  });

  it("loads and saves workspace state through authenticated endpoints", async () => {
    fetchMock
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            devices: [],
            conversations: [],
            activeConversationId: "",
            printJobs: [],
            schedules: [],
            sources: [],
            preferences: {
              loginProtectionEnabled: false,
              sendConfirmationEnabled: true,
              tutorialTabEnabled: true,
              theme: "light",
              defaultDeviceId: "",
            },
            serviceBinding: {
              providerName: null,
              modelName: "Ink AI",
              bound: false,
            },
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            devices: [],
            conversations: [],
            activeConversationId: "",
            printJobs: [],
            schedules: [],
            sources: [],
            preferences: {
              loginProtectionEnabled: false,
              sendConfirmationEnabled: true,
              tutorialTabEnabled: true,
              theme: "light",
              defaultDeviceId: "",
            },
            serviceBinding: {
              providerName: null,
              modelName: "Ink AI",
              bound: false,
            },
          }),
          { status: 200 },
        ),
      );

    await expect(fetchWorkspaceStateWithApi("access-token")).resolves.toMatchObject({
      preferences: {
        theme: "light",
      },
    });

    await expect(
      saveWorkspaceStateWithApi("access-token", {
        devices: [],
        conversations: [],
        activeConversationId: "",
        printJobs: [],
        schedules: [],
        sources: [],
        preferences: {
          loginProtectionEnabled: false,
          sendConfirmationEnabled: true,
          tutorialTabEnabled: true,
          theme: "light",
          defaultDeviceId: "",
        },
        serviceBinding: {
          providerName: null,
          modelName: "Ink AI",
          bound: false,
        },
      }),
    ).resolves.toMatchObject({
      serviceBinding: {
        modelName: "Ink AI",
      },
    });
  });

  it("creates users through the admin endpoint", async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          user: {
            id: "user-2",
            email: "new-user",
            name: "New User",
            role: "member",
          },
        }),
        { status: 201 },
      ),
    );

    await expect(
      createUserWithApi("access-token", {
        email: "new-user",
        name: "New User",
        password: "demo-password",
      }),
    ).resolves.toMatchObject({
      email: "new-user",
      role: "member",
    });
  });
});
