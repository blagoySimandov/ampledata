Use as little code as possible make it pretty. use tailwind. if creating new colors add them to

> tailwind config. use kebab case for files put constants in constants file types in types file
> functions in functions file. always export via index.ts file via a folder and not through the file inside the folder.

When creating or refactoring components always divide into small components
prefferbly in different files unless those compoenents are highly related and very small. If components to relate only to a certain component
put the main component as a folder name all smaller components in the same folder.
and the main "combination" of them in the index.ts file of that folder.

if a component is used in multiple places or you feel like it might be used in the future
put it under the src/components folder.
