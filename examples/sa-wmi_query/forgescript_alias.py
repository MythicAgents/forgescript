import forgescript

def wmi_query(task: forgescript.Task) -> forgescript.AliasedCommand:
    resource = f"\\\\{task.args['system']}\\{task.args['namespace']}"
    return forgescript.AliasedCommand("execute_coff", args={
        "bof_file": forgescript.register_file(f"bin/wmi_query.{task.callback.architecture}.o"),
        "coff_arguments": [
            ["string", task.args["system"]],
            ["string", task.args["namespace"]],
            ["string", task.args["query"]],
            ["string", resource]
        ]
    })

forgescript.register_alias("sa-wmi_query", wmi_query,
    description="Runs TrustedSec's wmi_query Situational Awareness BOF",
    author="TrustedSec",
    parameters=[
        forgescript.AliasParameter(
            "query",
            display_name="Query",
            description="Query to run. This should be in WQL",
            type=forgescript.AliasParameterType.String,
        ),
        forgescript.AliasParameter(
            "system",
            display_name="System",
            description="Remote system to connect to",
            type=forgescript.AliasParameterType.String,
            default_value="."
        ),
        forgescript.AliasParameter(
            "namespace",
            display_name="Namespace",
            description="Namespace to connect to",
            type=forgescript.AliasParameterType.String,
            default_value="root\\cimv2"
        )
    ]
)
