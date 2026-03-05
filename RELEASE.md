# Release Guide

This document explains how to build and release Agent Orchestrator for macOS, Windows, and Linux.

---

## 🚀 Quick Release

### Using GitHub Actions (Recommended)

1. **Tag and Push:**
```bash
# Update version in package.json first
# Then create and push a tag
git tag v0.1.0
git push origin v0.1.0
```

2. **GitHub Actions will:**
   - Build backend for all platforms
   - Build Electron app for macOS, Windows, Linux
   - Create a GitHub Release
   - Upload all artifacts

3. **Download from:**
   - https://github.com/formatho/agent-orchestrator/releases

---

## 🛠️ Local Build

### Prerequisites
- Go 1.24+
- Node.js 18+
- npm

### Build for Current Platform

```bash
cd electron-app
npm run build:electron
```

### Build for Specific Platform

**macOS:**
```bash
npm run build:mac
```
Output: `release/Agent-Orchestrator-{version}-mac.dmg`

**Windows:**
```bash
npm run build:win
```
Output: `release/Agent-Orchestrator-Setup-{version}.exe`

**Linux:**
```bash
npm run build:linux
```
Output: `release/Agent-Orchestrator-{version}.AppImage`

**All Platforms:**
```bash
npm run build:all
```

---

## 📦 Build Outputs

### macOS
- **DMG** - Disk image for easy installation
- **ZIP** - For auto-updates
- Architectures: x64 (Intel), arm64 (Apple Silicon)

### Windows
- **NSIS Installer** - Standard Windows installer
- **Portable** - Standalone executable
- Architectures: x64, ia32

### Linux
- **AppImage** - Universal Linux package
- **DEB** - Debian/Ubuntu package
- **RPM** - Red Hat/Fedora package

---

## 🏗️ Build Architecture

### What Gets Built

1. **Backend (Go Binary)**
   - Compiled for each platform
   - Embedded in Electron app resources
   - Runs automatically when app starts

2. **Frontend (Electron App)**
   - React + TypeScript bundle
   - Electron main process
   - Platform-specific packaging

3. **Complete Package**
   - Self-contained application
   - No external dependencies
   - SQLite database included

---

## 🔐 Code Signing (Production)

### macOS
1. **Apple Developer Certificate required**
2. **Notarization required for distribution**
3. Update `build/entitlements.mac.plist` if needed

### Windows
1. **Code signing certificate required**
2. Add certificate to GitHub Secrets
3. Update workflow to sign executable

---

## 📝 Release Checklist

Before releasing:

- [ ] Update version in `package.json`
- [ ] Update `CHANGELOG.md`
- [ ] Test build locally
- [ ] Verify all features work
- [ ] Check backend integration
- [ ] Test on target platforms (if possible)
- [ ] Create and push git tag
- [ ] Wait for GitHub Actions
- [ ] Verify release artifacts
- [ ] Test download and installation
- [ ] Announce release

---

## 🎯 Version Numbering

Use semantic versioning: `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes
- **MINOR**: New features
- **PATCH**: Bug fixes

Examples:
- `0.1.0` - Initial beta
- `0.2.0` - New features added
- `0.1.1` - Bug fixes
- `1.0.0` - First stable release

---

## 🔄 Auto-Updates

Agent Orchestrator supports auto-updates via electron-updater.

**Configuration:**
- Already configured in `package.json`
- Uses GitHub Releases as update server
- Checks for updates automatically

**How it works:**
1. App checks GitHub Releases on startup
2. If new version found, downloads in background
3. Prompts user to install
4. Updates on next launch

---

## 📊 Build Sizes (Approximate)

| Platform | Size |
|----------|------|
| macOS (DMG) | ~150MB |
| Windows (EXE) | ~140MB |
| Linux (AppImage) | ~160MB |

---

## 🐛 Troubleshooting

### Build Fails
1. Clean node_modules: `rm -rf node_modules && npm ci`
2. Check Go version: `go version`
3. Check Node version: `node -v`
4. Check platform-specific dependencies

### Code Signing Fails
1. Verify certificate is valid
2. Check Apple Developer account
3. Ensure provisioning profiles are correct

### GitHub Actions Fails
1. Check workflow logs
2. Verify all secrets are set
3. Check artifact upload/download

---

## 🚢 Distribution

### GitHub Releases (Free)
- Automatic with GitHub Actions
- No cost
- Public releases

### Mac App Store (Paid)
- Requires Apple Developer account ($99/year)
- App Store review required
- Sandboxing restrictions apply

### Microsoft Store (Paid)
- Requires Microsoft Developer account ($19 one-time)
- Store certification required

### Direct Distribution (Free)
- Host on your own server
- Use CDN for faster downloads
- Manage updates yourself

---

## 📧 Support

For build issues:
- Open an issue on GitHub
- Include build logs
- Specify platform and version

---

**Built with ❤️ by Formatho**
