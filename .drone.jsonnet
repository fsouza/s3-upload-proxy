// the first version is used to build the binary that gets shipped to Docker Hub.
local go_versions = ['1.12.7', '1.11.12', '1.13beta1'];

local test_dockerfile = {
  name: 'test-dockerfile',
  image: 'plugins/docker',
  settings: {
    repo: 'fsouza/s3-upload-proxy',
    dry_run: true,
  },
  when: {
    event: ['pull_request'],
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
    'goreleaser release',
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

local release_steps = [
  test_dockerfile,
  push_to_dockerhub,
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
  commands: ['golangci-lint run --enable-all -D lll -D errcheck ./...'],
  depends_on: ['mod-download'],
};

local build(go_version) = {
  name: 'build',
  image: 'golang:%(go_version)s' % { go_version: go_version },
  commands: ['go build -o s3-upload-proxy -mod readonly'],
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
  name: 'build_go_%(go_version)s' % { go_version: go_version },
  workspace: {
    base: '/go',
    path: 's3-upload-proxy-%(go_version)s' % { go_version: go_version },
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
