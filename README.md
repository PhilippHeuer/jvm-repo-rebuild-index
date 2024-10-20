# Index for Reproducible Builds / jvm-repo-rebuild

> [reproducible-central](https://github.com/jvm-repo-rebuild/reproducible-central) is an effort to rebuild public releases published to Maven Central and check that Reproducible Build can be achieved.

This project aims to provide a machine-readable index to query the reproducibility status of projects and artifacts.
Additionally, it provides dynamic badge endpoints for initial integration in tools like Renovate, Dependabot, etc.

> **Note:** This project is in early development and the format of the index may be updated, but the api endpoints will remain stable.

## Status

The index is updated automatically using GitHub Actions and published to GitHub Pages (once every 6 hours).

Note: If any files of a given project version are not reproducible, the entire project version including all artifacts will be reported as not reproducible.

## Usage

The generated index files are hosted on GitHub Pages and can be accessed using the following URLs:

| URL                                                                                                            | Description                                                |
|----------------------------------------------------------------------------------------------------------------|------------------------------------------------------------|
| `https://philippheuer.github.io/jvm-repo-rebuild-index/index.json`                                             | All maven repositories                                     |
| `https://philippheuer.github.io/jvm-repo-rebuild-index/mavencentral/project/{group}/{artifact}/index.json`     | Query project data                                         |
| `https://philippheuer.github.io/jvm-repo-rebuild-index/mavencentral/project/{group}/{artifact}/{version}.json` | Query project by group, artifact and version               |
| `https://philippheuer.github.io/jvm-repo-rebuild-index/mavencentral/maven/index.json`                          | All artifacts (currently disabled, due to large file size) |
| `https://philippheuer.github.io/jvm-repo-rebuild-index/mavencentral/maven/{group}/{artifact}/index.json`       | Query artifacts by group and artifact                      |
| `https://philippheuer.github.io/jvm-repo-rebuild-index/mavencentral/maven/{group}/{artifact}/{version}.json`   | Query artifacts by group, artifact and version             |

**Note**: Replace all dots (`.`) in the group with slashes (`/`) in the URL.

## Example Queries

**Note**: The project files follow the structure of `reproducible-central`, while the artifact files are generated based on the individual maven coordinates.

```bash
# project - io.github.xanthic.cache:cache-api
https://philippheuer.github.io/jvm-repo-rebuild-index/mavencentral/project/io/github/xanthic/cache/cache-api/index.json
# project - io.github.xanthic.cache:cache-api:0.6.2
https://philippheuer.github.io/jvm-repo-rebuild-index/mavencentral/project/io/github/xanthic/cache/cache-api/0.6.2.json

# artifact - io.github.xanthic.cache.cache-provider-caffeine3
https://philippheuer.github.io/jvm-repo-rebuild-index/mavencentral/maven/io/github/xanthic/cache/cache-provider-caffeine3/index.json
# artifact - io.github.xanthic.cache.cache-provider-caffeine3:0.6.2
https://philippheuer.github.io/jvm-repo-rebuild-index/mavencentral/maven/io/github/xanthic/cache/cache-provider-caffeine3/0.6.2.json
```

## Badges

You can use the `Endpoint Badge` of shields.io to display the reproducibility status of a project, artifact or dependencies.

### Project Badge

```markdown
# project - io.github.xanthic.cache:cache-api (all artifacts)
![Reproducible Builds](https://img.shields.io/endpoint?url=https://jvm-rebuild.philippheuer.de/v1/badge/reproducible/project/io.github.xanthic.cache:cache-api/latest)
```

![Reproducible Builds](https://img.shields.io/endpoint?url=https://jvm-rebuild.philippheuer.de/v1/badge/reproducible/project/io.github.xanthic.cache:cache-api/latest)

### Artifact Badge

```markdown
# artifact - io.github.xanthic.cache:cache-provider-caffeine3
![Reproducible Builds](https://img.shields.io/endpoint?url=https://jvm-rebuild.philippheuer.de/v1/badge/reproducible/maven/io.github.xanthic.cache:cache-api/latest)
```

![Reproducible Builds](https://img.shields.io/endpoint?url=https://jvm-rebuild.philippheuer.de/v1/badge/reproducible/maven/io.github.xanthic.cache:cache-api/latest)

### Dependency Badge (Experimental)

The dependency badge counts all dependencies of a library that are reproducible.

**Note**: This badge is experimental and could break at any time, as it's using unofficial endpoints.

```markdown
# artifact - io.github.xanthic.cache:cache-provider-caffeine3
![Reproducible Builds](https://img.shields.io/endpoint?url=https://jvm-rebuild.philippheuer.de/v1/badge/reproducible-dependencies/maven/io.github.xanthic.cache:cache-provider-cache2k/0.6.2)
```

![Reproducible Builds](https://img.shields.io/endpoint?url=https://jvm-rebuild.philippheuer.de/v1/badge/reproducible-dependencies/maven/io.github.xanthic.cache:cache-provider-cache2k/0.6.2)

## License

The code is released under the [MIT license](./LICENSE).

The included logos are trademarks of their respective owners and are not covered by the license.
