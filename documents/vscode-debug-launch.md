# VS Code Debug Launch Setup

This document shows how to add a `launch.json` configuration for SimonOS so you can debug the `run` command in VS Code.

## Sample `launch.json`

Create or update `.vscode/launch.json` with this content:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug SimonOS: run (launch)",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/simonos",
      "cwd": "${workspaceFolder}",
      "args": [
        "run",
        "--config",
        "${workspaceFolder}/configs/config.example.yaml",
        "--input",
        "tell me about napoleon"
      ],
      "env": {
        "SIMONOS_MODEL_API_TYPE": "ollama",
        "SIMONOS_MODEL_HOST": "http://localhost:11434",
        "SIMONOS_MODEL_ID": "lfm2:24b"
      },
      "buildFlags": "",
      "showLog": false
    },
    {
      "name": "Attach to Delve (localhost:2345)",
      "type": "go",
      "request": "attach",
      "mode": "remote",
      "remotePath": "",
      "port": 2345,
      "host": "127.0.0.1",
      "showLog": false
    }
  ]
}
```

## How To Add It In VS Code

1. Open the SimonOS workspace in VS Code.
2. Open the Run and Debug view.
3. Click `create a launch.json file` if no launch configuration exists yet.
4. Choose the `Go` debugger when VS Code asks for an environment.
5. Replace the generated file contents with the sample above.
6. Save `.vscode/launch.json`.

## How To Use It

### Launch directly

1. Set breakpoints in files such as `cmd/simonos/run.go`, `internal/agent/engine.go`, or `internal/model/ollama_provider.go`.
2. In Run and Debug, select `Debug SimonOS: run (launch)`.
3. Press `F5`.

This starts SimonOS from the workspace root and loads the config file explicitly.

### Attach to a headless Delve server

Start Delve from a terminal:

```bash
dlv debug --headless --listen=:2345 --api-version=2 ./cmd/simonos -- run --input "tell me about napoleon"
```

Then in VS Code:

1. Open Run and Debug.
2. Select `Attach to Delve (localhost:2345)`.
3. Press `F5`.

## Notes

- The `cwd` entry is important because the config file path is otherwise resolved relative to the current working directory.
- The explicit `--config` argument avoids relative path issues during debugging.
- `showLog: false` keeps the terminal output focused on the program output instead of Delve logging.
- If Ollama is running on a different host or model, update the values in the `env` section.
