# Codeartoken

* CLI tool to refresh [AWS CodeArtifact](https://aws.amazon.com/codeartifact/) tokens for maven using the `settings.xml` file.

```shell
❯ ./codeartoken --help
NAME:
   codeartoken - Refresh AWS CodeArtifact token for maven

USAGE:
   codeartoken [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --domain value, -d value
   --owner value, -o value
   --server value, -s value    (default: "codeartifact")
   --settings value, -x value  (default: "$HOME/.m2/settings.xml")
   --help, -h                  show help (default: false)
```
