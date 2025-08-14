# sa-nslookup
This example will create an alias command `sa-nslookup` which will run the TrustedSec `nslookup`
BOF from the [TrustedSec CS Situational Awareness BOF](https://github.com/trustedsec/CS-Situational-Awareness-BOF) repository.


## Dependencies
Requires placing the compiled [`nslookup.x64.o`](https://github.com/trustedsec/CS-Situational-Awareness-BOF/tree/master/SA/nslookup) and [`nslookup.x86.o`](https://github.com/trustedsec/CS-Situational-Awareness-BOF/tree/master/SA/nslookup) BOFs in the `bin/` directory.
```
bin/nslookup.x64.o
bin/nslookup.x86.o
```

## Building
```
make
```

## Usage
Upload the generated `sa-nslookup.tar.gz` file using the `forgescript_load` command.
