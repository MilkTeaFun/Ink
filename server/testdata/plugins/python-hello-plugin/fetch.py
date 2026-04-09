import json
import sys


def main() -> None:
    payload = json.loads(sys.stdin.read() or "{}")
    workspace_config = payload.get("workspaceConfig", {})
    schedule_config = payload.get("scheduleConfig", {})

    source_name = str(workspace_config.get("sourceName", "Python Hello Source")).strip()
    message = str(schedule_config.get("message", "Hello from the Python fixture plugin.")).strip()
    uppercase = bool(schedule_config.get("uppercase", False))

    content = message.upper() if uppercase else message
    sys.stdout.write(
        json.dumps(
            {
                "title": f"{source_name} Digest",
                "content": content,
                "sourceLabel": source_name,
            }
        )
    )


if __name__ == "__main__":
    main()
