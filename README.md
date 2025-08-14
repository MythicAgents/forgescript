# Forgescript
[![GitHub License](https://img.shields.io/github/license/MythicAgents/forgescript)](https://github.com/MythicAgents/forgescript/blob/main/LICENSE)
[![GitHub Release](https://img.shields.io/github/v/release/MythicAgents/forgescript)](https://github.com/MythicAgents/forgescript/releases/latest)
[![Release](https://github.com/MythicAgents/forgescript/workflows/Release/badge.svg)](https://github.com/MythicAgents/forgescript/actions/workflows/release.yml)

> [!WARNING]
> This project is currently in testing until version 0.1.0 is released.

Scriptable alias command augmentations in Mythic using Python.

## Installation
Install using the `mythic-cli` program on an existing Mythic system.
```bash
./mythic-cli install github https://github.com/MythicAgents/forgescript
```

## Features
Embedded scripting interface for registering custom command aliases with arbitrary agents.

### Example
An example script which creates an alias command `forgescript_whoami` which will run the builtin `whoami`
command for an agent.

```py
import forgescript


def whoami(task: forgescript.Task) -> forgescript.AliasedCommand:
    return forgescript.AliasedCommand("whoami")


forgescript.register_alias(
    "forgescript_whoami",
    whoami,
    description="Runs the builtin whoami command",
    author="MEhrn00",
)
```

More extended examples can be found in the [`examples/`](/examples/) directory.

## Commands
Command          | Syntax                     | Description
---------------- | -------------------------- | --------------------------------
forgescript_load | `forgescript_load [popup]` | Load a script bundle into Mythic
