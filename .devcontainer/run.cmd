@echo off
setlocal

set IMAGE_NAME=easi-claude-dev
set DEVCONTAINER_DIR=%~dp0.
set REPO_DIR=%~dp0..

if "%CS_ACCESS_TOKEN%"=="" (
  echo CS_ACCESS_TOKEN must be set before starting the development container.
  exit /b 1
)

if "%CS_ONPREM_URL%"=="" (
  echo CS_ONPREM_URL must be set before starting the development container.
  exit /b 1
)

podman build -t %IMAGE_NAME% -f "%DEVCONTAINER_DIR%\Dockerfile" "%DEVCONTAINER_DIR%" || exit /b 1

podman run -it --rm ^
  -v "%REPO_DIR%:/workspace:z" ^
  -v easi-claude-config:/home/node/.claude ^
  -v easi-go-cache:/home/node/go ^
  -v easi-npm-cache:/home/node/.npm ^
  -v easi-frontend-node-modules:/workspace/frontend/node_modules ^
  -e CS_ACCESS_TOKEN ^
  -e CS_ONPREM_URL ^
  --user node ^
  -w /workspace ^
  %IMAGE_NAME% sh -c "sh /workspace/.devcontainer/init.sh; exec zsh"
