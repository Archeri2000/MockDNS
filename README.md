# DNS Mock tool
This simple golang tool is written to perform simple DNS Mocking on various OS.

## Mocking DNS
DNS queries are first checked against an individual's hosts file before being sent out to DNS servers. 
As such, domain names can easily be mocked for the purposes of testing reverse proxies, websites and other network related services.


This command line tool requires **elevated administrator access**, and can be used to manage different mocked domain names, while minimising interference with the normal operation of the hosts dns file.

# How to install
Compile the golang files for the target system using the command:

`env GOOS=target-OS GOARCH=target-architecture go build package-import-path`

Currently supported OS are linux, darwin, windows.


If desired, use `go install MockDNS` to install the command in the PATH


# Commands

The following commands have been implemented:
*   `MockDNS add <Filename>` will add the dns entries in the file into the hosts file.
*  `MockDNS add-now <IP> <Domain Name>` will add the dns entry into the hosts file.
*  `MockDNS remove <Filename>` will remove the matching dns entries in the file from the hosts file.
*  `MockDNS remove-now <IP> <Domain Name>` will remove the matching dns entry from the hosts file.
*  `MockDNS reset` will remove all currently mocked DNS entries from the hosts file.
*  `MockDNS show` will display all currently mocked DNS entries.
*  `MockDNS show-all` will print the contents of the entire hosts file.

	
For more details on the format of the file used as the argument for the `add` and `remove` commands, please see Sample.txt.


## IP Addresses

This tool accepts both IPv4 and IPv6 addresses. Additionally, there is the following alias:
* `localhost` -> `127.0.0.1`
* `localhostv6` -> `::1`

	

## Misc
The changes made to the DNS are all tracked in a file called `changes` which is stored in the same directory that the executable exists in.

# Authors

*  Arch2K
