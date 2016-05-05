##Build Instructions

###Build 
- git tag -a v0.0.1.`git log --pretty=format:'%h' -n 1` -m 'RSVM v0.0.01.`git log --pretty=format:'%h' -n 1`'
- make
- make deb
