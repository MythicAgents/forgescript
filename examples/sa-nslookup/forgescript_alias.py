import forgescript

recordmap = {
    "A": 1,
    "NS": 2,
    "MD": 3,
    "MF": 4,
    "CNAME": 5,
    "SOA": 6,
    "MB": 7,
    "MG": 8,
    "MR": 9,
    "WKS": 0xb,
    "PTR": 0xc,
    "HINFO": 0xd,
    "MINFO": 0xe,
    "MX": 0xf,
    "TXT": 0x10,
    "RP": 0x11,
    "AFSDB": 0x12,
    "X25": 0x13,
    "ISDN": 0x14,
    "RT": 0x15,
    "AAAA": 0x1c,
    "SRV": 0x21,
    "WINSR": 0xff02,
    "KEY": 0x0019,
    "ANY": 0xff
}

def nslookup(task: forgescript.Task) -> forgescript.AliasedCommand:
    return forgescript.AliasedCommand("execute_coff", args={
        "bof_file": forgescript.register_file(f"bin/nslookup.{task.callback.architecture}.o"),
        "coff_arguments": [
            ["string", task.args["domain"]],
            ["string", task.args["server"]],
            ["int16", recordmap[task.args["type"]]],
        ]
    })

forgescript.register_alias("sa-nslookup", nslookup,
    description="Runs TrustedSec's nslookup Situational Awareness BOF",
    author="TrustedSec",
    parameters=[
        forgescript.AliasParameter(
            "domain",
            display_name="Domain name",
            description="Domain name to look up",
            type=forgescript.AliasParameterType.String,
        ),
        forgescript.AliasParameter(
            "server",
            display_name="DNS server",
            description="DNS server to query",
            type=forgescript.AliasParameterType.String,
            default_value=""
        ),
        forgescript.AliasParameter(
            "type",
            display_name="Record type",
            description="DNS record type to look up",
            type=forgescript.AliasParameterType.ChooseOne,
            choices=list(recordmap.keys()),
            default_value="A"
        )
    ]
)
