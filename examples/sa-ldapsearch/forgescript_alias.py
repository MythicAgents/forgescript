import forgescript

# More complicated ldapsearch example
def ldapsearch(task: forgescript.Task) -> forgescript.AliasedCommand:
    attributes = "*" if len(task.args["attributes"]) == 0 else ",".join(task.args["attributes"])
    ldaps = 1 if task.args["ldaps"] else 0

    scope = 3
    if task.args["scope"] == "BASE":
        scope = 1
    elif task.args["scope"] == "LEVEL":
        scope = 2
    elif task.args["scope"] == "SUBTREE":
        scope = 3
    else:
        raise RuntimeError(f"Invalid scope {task.args['scope']}")

    return forgescript.AliasedCommand("execute_coff", args={
        "bof_file": forgescript.register_file(f"bin/ldapsearch.{task.callback.architecture}.o"),
        "coff_arguments": [
            ["string", task.args["query"]],
            ["string", attributes],
            ["int32", task.args["count"]],
            ["int32", scope],
            ["string", task.args["hostname"]],
            ["string", task.args["dn"]],
            ["int32", ldaps]
        ]
    })

forgescript.register_alias("sa-ldapsearch", ldapsearch,
    description="Runs TrustedSec's ldapsearch Situational Awareness BOF",
    author="TrustedSec",
    parameters=[
    forgescript.AliasParameter(
        "query",
        display_name="LDAP query to perform",
        description="LDAP query to perfrom",
        type=forgescript.AliasParameterType.String,
    ),
    forgescript.AliasParameter(
        "attributes",
        description="the attributes to retrieve",
        type=forgescript.AliasParameterType.Array,
        default_value=[],
    ),
    forgescript.AliasParameter(
        "count",
        display_name="Maximum results to return",
        description="the result max size",
        type=forgescript.AliasParameterType.Number,
        default_value=0,
    ),
    forgescript.AliasParameter(
        "scope",
        display_name="Query scope",
        description="the scope to use",
        type=forgescript.AliasParameterType.ChooseOne,
        choices=["BASE", "LEVEL", "SUBTREE"],
        default_value="SUBTREE",
    ),
    forgescript.AliasParameter(
        "hostname",
        description="hostname or IP to perform the LDAP connection on (default: automatic DC resolution)",
        type=forgescript.AliasParameterType.String,
        default_value="",
    ),
    forgescript.AliasParameter(
        "dn",
        display_name="the LDAP query base",
        type=forgescript.AliasParameterType.String,
        default_value="",
    ),
    forgescript.AliasParameter(
        "ldaps",
        description="use of ldaps",
        type=forgescript.AliasParameterType.Boolean,
        default_value=False,
    ),
])
