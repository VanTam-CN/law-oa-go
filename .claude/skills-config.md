# Claude Skills Configuration

## Skills Management

This file configures the skills system for Claude Code.

### Available Skills

Currently, the following skills are available:

1. **Building Native UI** - Complete guide for building beautiful apps with Expo Router
2. **Native Data Fetching** - Use when implementing or debugging ANY network request, API call, or data fetching
3. **Remotion Best Practices** - Best practices for Remotion - Video creation in React
4. **Skill Creator** - Guide for creating effective skills
5. **Upgrading Expo** - Guidelines for upgrading Expo SDK versions and fixing dependency issues
6. **Vercel React Best Practices** - React and Next.js performance optimization guidelines from Vercel Engineering
7. **Web Design Guidelines** - Review UI code for Web Interface Guidelines compliance

### Skills Directory

Skills are stored in the `/tmp/` directory and managed through the Claude Code skills system.

### Configuration

To enable/disable specific skills, modify this file.

### Skills Server

The skills server is managed by Claude Code and should be available when needed.

## Notes

- Skills are loaded dynamically when needed
- Each skill has specific use cases and should be called when appropriate
- Skills can provide specialized assistance for complex tasks