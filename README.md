# Reproducible Central Index

> [reproducible-central](https://github.com/jvm-repo-rebuild/reproducible-central) is an effort to rebuild public releases published to Maven Central and check that Reproducible Build can be achieved.
> This projects aims to provide a machine-readable index for third party tools to look up the reproducibility status of a given artifact by group/artifact/version.

This index is not affiliated with the [Reproducible Builds](https://reproducible-builds.org/) project, but aims to provide a machine-readable index to query the reproducibility status of maven artifacts.
This should help with the integration of tools like dependabot, renovate as part of dependency management.

## Status

The index is updated automatically using GitHub Actions and published to GitHub Pages.

Note: If any files of a given project version are not reproducible, the entire project version including all artifacts will be reported as not reproducible.

## Usage

The generated index files are hosted on GitHub Pages and can be accessed using the following URLs:

| URL                                                                                                           | Description                                                |
|---------------------------------------------------------------------------------------------------------------|------------------------------------------------------------|
| `https://philippheuer.github.io/reproducible-central-index/index.json`                                        | All maven repositories                                     |
| `https://philippheuer.github.io/reproducible-central-index/central/project/{group}/{artifact}/index.json`     | Query project data                                         |
| `https://philippheuer.github.io/reproducible-central-index/central/project/{group}/{artifact}/{version}.json` | Query project by group, artifact and version               |
| `https://philippheuer.github.io/reproducible-central-index/central/project/{group}/{artifact}/badge.json`     | Project shields.io badge endpoint                          |
| `https://philippheuer.github.io/reproducible-central-index/central/maven/index.json`                          | All artifacts (currently disabled, due to large file size) |
| `https://philippheuer.github.io/reproducible-central-index/central/maven/{group}/{artifact}/index.json`       | Query artifacts by group and artifact                      |
| `https://philippheuer.github.io/reproducible-central-index/central/maven/{group}/{artifact}/{version}.json`   | Query artifacts by group, artifact and version             |
| `https://philippheuer.github.io/reproducible-central-index/central/maven/{group}/{artifact}/badge.json`       | Artifact shields.io badge endpoint                         |

## Examples

```bash
# project - io.github.xanthic.cache:cache-api
https://philippheuer.github.io/reproducible-central-index/central/project/io/github/xanthic/cache/cache-api/index.json
# project - io.github.xanthic.cache:cache-api:0.6.2
https://philippheuer.github.io/reproducible-central-index/central/project/io/github/xanthic/cache/cache-api/0.6.2.json

# artifact - io.github.xanthic.cache.cache-provider-caffeine3
https://philippheuer.github.io/reproducible-central-index/central/maven/io/github/xanthic/cache/cache-provider-caffeine3/index.json
# artifact - io.github.xanthic.cache.cache-provider-caffeine3:0.6.2
https://philippheuer.github.io/reproducible-central-index/central/maven/io/github/xanthic/cache/cache-provider-caffeine3/0.6.2.json
```

## Badges

You can use the `Endpoint Badge` of shields.io to display the reproducibility status of the latest release:

```markdown
# project - io.github.xanthic.cache:cache-api (all artifacts)
![Reproducible Builds](https://img.shields.io/endpoint?url=https://philippheuer.github.io/reproducible-central-index/central/project/io/github/xanthic/cache/cache-api/badge.json)
```

![Reproducible Builds](https://img.shields.io/endpoint?url=https://philippheuer.github.io/reproducible-central-index/central/project/io/github/xanthic/cache/cache-api/badge.json)

```markdown
# artifact - io.github.xanthic.cache:cache-provider-caffeine3
![Reproducible Builds](https://img.shields.io/endpoint?url=https://philippheuer.github.io/reproducible-central-index/central/maven/io/github/xanthic/cache/cache-provider-caffeine3/badge.json)
```

![Reproducible Builds](https://img.shields.io/endpoint?url=https://philippheuer.github.io/reproducible-central-index/central/maven/io/github/xanthic/cache/cache-provider-caffeine3/badge.json)

## FAQ

### How often is the index updated?

The index is updated automatically using GitHub Actions and published to GitHub Pages. This workflow runs every 12 hours.

### The latest version is missing / not up-to-date

If the version schema of the artifact is not semver compliant, the latest version may be reported falsely as we use a semver comparison to determine the latest version.

### Why is an artifact marked as non-reproducible even though all files of the artifact are reproducible?

Projects often have internal dependencies on other modules of the same project.
Therefore, the index reports all artifacts of a project version as non-reproducible if any of the artifacts are not reproducible.

## License

The code is released under the [MIT license](./LICENSE).
