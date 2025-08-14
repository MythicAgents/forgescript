import forgescript


def whoami(task: forgescript.Task) -> forgescript.AliasedCommand:
    return forgescript.AliasedCommand("execute_coff", args={
        "bof_file": forgescript.register_file(f"bin/whoami.{task.callback.architecture}.o")
    })


forgescript.register_alias(
    "sa-whoami",
    whoami,
    description="Runs TrustedSec's whoami Situational Awareness BOF",
    author="TrustedSec",
)
