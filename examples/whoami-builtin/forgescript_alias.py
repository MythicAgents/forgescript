import forgescript


def whoami(task: forgescript.Task) -> forgescript.AliasedCommand:
    return forgescript.AliasedCommand("whoami")


forgescript.register_alias(
    "forgescript_whoami",
    whoami,
    description="Runs the builtin whoami command",
)
