#!/bin/bash

# Build Script for Agent Orchestrator
# Builds backend (Go) and frontend (Electron) for distribution

set -e

VERSION=${1:-"0.1.0"}
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
BACKEND_DIR="$PROJECT_ROOT/backend"
ELECTRON_DIR="$PROJECT_ROOT/electron-app"
DIST_DIR="$PROJECT_ROOT/dist"

echo "🏗️  Building Agent Orchestrator v$VERSION"
echo "============================================"
echo ""

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# Build Backend (Go)
echo ""
echo "📦 Building Backend (Go)..."
cd "$BACKEND_DIR"

# Build for multiple platforms
platforms=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "windows/amd64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"

    output_name="server-$GOOS-$GOARCH"
    if [ "$GOOS" = "windows" ]; then
        output_name="$output_name.exe"
    fi

    echo "  Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w -X main.Version=$VERSION" -o "bin/$output_name" ./cmd/server
done

echo "✅ Backend builds complete"

# Build Frontend (Electron)
echo ""
echo "🎨 Building Frontend (Electron)..."
cd "$ELECTRON_DIR"

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "  Installing dependencies..."
    npm ci
fi

# Build for current platform
echo "  Building Electron app..."
npm run build:electron

# Move releases to dist folder
if [ -d "release" ]; then
    echo "  Moving releases to dist..."
    cp -r release/* "$DIST_DIR/"
fi

echo ""
echo "✅ Build Complete!"
echo ""
echo "📂 Output Location: $DIST_DIR"
echo ""
echo "📊 Built Artifacts:"
ls -lh "$DIST_DIR" 2>/dev/null || echo "  (No artifacts yet - run with platform targets)"
echo ""
echo "🚀 Build commands:"
echo "  ./scripts/build.sh           - Build for current platform"
echo "  npm run build:mac            - Build for macOS (DMG + ZIP)"
echo "  npm run build:win            - Build for Windows (EXE)"
echo "  npm run build:linux          - Build for Linux (AppImage + DEB)"
echo "  npm run build:all            - Build for all platforms"
echo ""
echo "🎯 To release:"
echo "  git tag v$VERSION"
echo "  git push origin v$VERSION"
echo "  # GitHub Actions will build and create release automatically"
