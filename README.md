## To be updated

can 
- configure username
- track files for any changes and stage them
- log commits
- commit changes
- retrieve files from a specific commit

 ```text
 when files being tracked are modified and commited,
 it basically creates unique commit directories, and save files to it.
```
  // comparision algo: 

create individual hash of all staged files, and then hash the individual hashes to generate a new commit hash.
if the commit hash is same as just the prev commit ID -> No files were changed since the last commmit, so abort.
else, make another commit.

- Monitoring bug ( now fixed) : 
https://github.com/vishalvivekm/vcs/issues/3
