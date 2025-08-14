# sa-wmi_query
This example will create an alias command `sa-wmi_query` which will run the TrustedSec `wmi_query`
BOF from the [TrustedSec CS Situational Awareness BOF](https://github.com/trustedsec/CS-Situational-Awareness-BOF) repository.


## Dependencies
Requires placing the compiled [`wmi_query.x64.o`](https://github.com/trustedsec/CS-Situational-Awareness-BOF/tree/master/SA/wmi_query) and [`wmi_query.x86.o`](https://github.com/trustedsec/CS-Situational-Awareness-BOF/tree/master/SA/wmi_query) BOFs in the `bin/` directory.
```
bin/wmi_query.x64.o
bin/wmi_query.x86.o
```

## Building
```
make
```

## Usage
Upload the generated `sa-wmi_query.tar.gz` file using the `forgescript_load` command.
