export default {
  'frontend/**/*.{ts,vue}': ['eslint --fix', 'prettier --write'],
  'frontend/**/*.{css,json,md}': ['prettier --write'],
  'backend/**/*.go': ['golangci-lint run --fix --new-from-rev=HEAD'],
}
