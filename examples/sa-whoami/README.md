# sa-whoami
This example will create an alias command `sa-whoami` which will run the TrustedSec `whoami`
BOF from the [TrustedSec CS Situational Awareness BOF](https://github.com/trustedsec/CS-Situational-Awareness-BOF) repository.


## Dependencies
Requires placing the compiled [`whoami.x64.o`](https://github.com/trustedsec/CS-Situational-Awareness-BOF/tree/master/SA/whoami) and [`whoami.x86.o`](https://github.com/trustedsec/CS-Situational-Awareness-BOF/tree/master/SA/whoami) BOFs in the `bin/` directory.
```
bin/whoami.x64.o
bin/whoami.x86.o
```

## Building
```
make
```

## Usage
Upload the generated `sa-whoami.tar.gz` file using the `forgescript_load` command.
