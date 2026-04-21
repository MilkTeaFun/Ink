import json
import sys


def main() -> None:
    payload = json.loads(sys.stdin.read() or "{}")
    workspace_config = payload.get("workspaceConfig", {})
    trigger = payload.get("trigger", {}) or {}

    source_name = str(workspace_config.get("sourceName", "Python Hello Source")).strip()
    message = str(workspace_config.get("message", "Hello from the Python fixture plugin.")).strip()
    uppercase = bool(workspace_config.get("uppercase", False))

    body = message.upper() if uppercase else message
    external_id = f"python-hello-{trigger.get('triggeredAt', 'default')}"
    item = {
        "externalId": external_id,
        "title": f"{source_name} Digest",
        "sourceLabel": source_name,
        "blocks": [
            {"type": "heading", "level": 1, "text": f"{source_name} Digest"},
            {"type": "paragraph", "text": body},
        ],
    }

    sys.stdout.write(
        json.dumps(
            {
                "items": [item],
                "cursor": trigger.get("triggeredAt"),
            }
        )
    )


if __name__ == "__main__":
    main()
