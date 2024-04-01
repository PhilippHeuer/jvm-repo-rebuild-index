# Reproducible Central Index

> [reproducible-central](https://github.com/jvm-repo-rebuild/reproducible-central) is an effort to rebuild public releases published to Maven Central and check that Reproducible Build can be achieved.
> This projects aims to provide a machine-readable index for third party tools to look up the reproducibility status of a given artifact by group/artifact/version.

This index is not affiliated with the [Reproducible Builds](https://reproducible-builds.org/) project and a temporary solution until there is a official way to query the status of a given artifact.

## Status

The index is updated once per hour using GitHub Actions and published to GitHub Pages.

## Usage

The generated index files are hosted on GitHub Pages and can be accessed using the following URLs:

| URL                                                                                    | Description                    |
|----------------------------------------------------------------------------------------|--------------------------------|
| `https://philippheuer.me/reproducible-central-index/index.json`                        | All artifacts                  |
| `https://philippheuer.me/reproducible-central-index/{group}/{artifact}/index.json`     | By group and artifact          |
| `https://philippheuer.me/reproducible-central-index/{group}/{artifact}/{version}.json` | By group, artifact and version |
| `https://philippheuer.me/reproducible-central-index/{group}/{artifact}/badge.json`     | for shields.io endpoint badge  |

_Examples:_

com.fasterxml.jackson.core:jackson-databind

`curl https://philippheuer.me/reproducible-central-index/com/fasterxml/jackson/core/jackson-databind/index.json`

com.fasterxml.jackson.core:jackson-databind:2.17.0

`curl https://philippheuer.me/reproducible-central-index/com/fasterxml/jackson/core/jackson-databind/2.17.0.json`

## Badges

You can use the `Endpoint Badge` of shields.io to display the reproducibility status of the latest release:

```markdown
# example for com.fasterxml.jackson.core:jackson-databind
![Reproducible Builds](https://img.shields.io/endpoint?url=https%3A%2F%2Fphilippheuer.me%2Freproducible-central-index%2Fcom%2Ffasterxml%2Fjackson%2Fcore%2Fjackson-databind%2Fbadge.json)
```

Demo:

![Reproducible Builds](https://img.shields.io/endpoint?url=https%3A%2F%2Fphilippheuer.me%2Freproducible-central-index%2Fcom%2Ffasterxml%2Fjackson%2Fcore%2Fjackson-databind%2Fbadge.json)

## License

The code is released under the [MIT license](./LICENSE).
