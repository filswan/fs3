# fs3

## Installation dependency

Run `npm install` to generate component.

## Development server

Run `npm run dev` for a dev server. Navigate to `http://localhost:8080/`. The app will automatically reload if you change any of the source files.

## Build project

```shell
# install cross-env
$ npm i cross-env --save

# Build test projects
$ npm run release:test

# Build calibration projects
$ npm run release:calibration

# Build production projects
$ npm run release:prod
```

The build artifacts will be stored in the `release/` directory.

## Further help

For a detailed explanation on how things work, check out the [guide](http://vuejs-templates.github.io/webpack/) and [docs for vue-loader](http://vuejs.github.io/vue-loader).
