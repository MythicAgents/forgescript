# sa-ldapsearch
This example will create an alias command `sa-ldapsearch` which will run the TrustedSec `ldapsearch`
BOF from the [TrustedSec CS Situational Awareness BOF](https://github.com/trustedsec/CS-Situational-Awareness-BOF) repository.


## Dependencies
Requires placing the compiled [`ldapsearch.x64.o`](https://github.com/trustedsec/CS-Situational-Awareness-BOF/tree/master/SA/ldapsearch) and [`ldapsearch.x86.o`](https://github.com/trustedsec/CS-Situational-Awareness-BOF/tree/master/SA/ldapsearch) BOFs in the `bin/` directory.
```
bin/ldapsearch.x64.o
bin/ldapsearch.x86.o
```

## Building
```
make
```

## Usage
Upload the generated `sa-ldapsearch.tar.gz` file using the `forgescript_load` command.
