// the first version is used to build the binary that gets shipped to Docker Hub.
local go_versions = ['1.13.1', '1.12.10'];

local test_dockerfile = {
  name: 'test-dockerfile',
  image: 'plugins/docker',
  settings: {
    repo: 'fsouza/s3-upload-proxy',
    dry_run: true,
    build_args: [
      'GOPROXY=https://proxy.golang.org',
    ],
  },
  when: {
    event: ['push', 'pull_request'],
  },
  depends_on: ['clone'],
};

local push_to_dockerhub = {
  name: 'build-and-push-to-dockerhub',
  image: 'plugins/docker',
  settings: {
    repo: 'fsouza/s3-upload-proxy',
    auto_tag: true,
    dockerfile: 'ci/Dockerfile',
    username: { from_secret: 'docker_user' },
    password: { from_secret: 'docker_password' },
  },
  when: {
    ref: [
      'refs/tags/*',
      'refs/heads/master',
    ],
  },
  depends_on: ['test', 'lint', 'build'],
};

local goreleaser = {
  name: 'goreleaser',
  image: 'goreleaser/goreleaser',
  commands: [
    'git fetch --tags',
    'goreleaser release -f ci/.goreleaser.yml',
  ],
  environment: {
    GITHUB_TOKEN: {
      from_secret: 'github_token',
    },
  },
  depends_on: ['test', 'lint'],
  when: {
    event: ['tag'],
  },
};

local goreleaser_test = {
  name: 'test-goreleaser',
  image: 'goreleaser/goreleaser',
  commands: [
    'goreleaser release --snapshot -f ci/.goreleaser.yml',
  ],
  depends_on: ['clone'],
  when: {
    event: ['push', 'pull_request'],
  },
};

local release_steps = [
  test_dockerfile,
  push_to_dockerhub,
  goreleaser_test,
  goreleaser,
];

local mod_download(go_version) = {
  name: 'mod-download',
  image: 'golang:%(go_version)s' % { go_version: go_version },
  commands: ['go mod download'],
  environment: { GOPROXY: 'https://proxy.golang.org' },
  depends_on: ['clone'],
};

local tests(go_version) = {
  name: 'test',
  image: 'golang:%(go_version)s' % { go_version: go_version },
  commands: ['go test -race -vet all -mod readonly ./...'],
  depends_on: ['mod-download'],
};

local lint = {
  name: 'lint',
  image: 'golangci/golangci-lint',
  pull: 'always',
  commands: ['golangci-lint run'],
  depends_on: ['mod-download'],
};

local build(go_version) = {
  name: 'build',
  image: 'golang:%(go_version)s' % { go_version: go_version },
  commands: ['go build -o s3-upload-proxy -mod readonly'],
  environment: { CGO_ENABLED: '0' },
  depends_on: ['mod-download'],
};

local test_ci_dockerfile = {
  name: 'test-ci-dockerfile',
  image: 'plugins/docker',
  settings: {
    repo: 'fsouza/s3-upload-proxy',
    dockerfile: 'ci/Dockerfile',
    dry_run: true,
  },
  when: {
    event: ['pull_request'],
  },
  depends_on: ['build'],
};

local pipeline(go_version) = {
  kind: 'pipeline',
  name: 'go:%(go_version)s' % { go_version: go_version },
  workspace: {
    base: '/go',
    path: 's3-upload-proxy',
  },
  steps: [
    mod_download(go_version),
    tests(go_version),
    lint,
    build(go_version),
    test_ci_dockerfile,
  ] + if go_version == go_versions[0] then release_steps else [],
};

std.map(pipeline, go_versions)
