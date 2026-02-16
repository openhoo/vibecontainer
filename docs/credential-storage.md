# Credential Storage

The `vibecontainer` CLI securely stores OAuth tokens and API keys in your system keychain, eliminating the need to re-enter credentials every time you create a stack.

## Supported Credentials

The following credentials are automatically stored:

- **Claude OAuth Token** - For Claude Code authentication
- **Anthropic API Key** - For Claude API authentication
- **Codex Auth JSON** - Full Codex authentication payload
- **OpenAI API Key** - For OpenAI/Codex API authentication
- **Codex API Key** - Codex-specific API key
- **Tunnel Token** - Cloudflare tunnel token

## How It Works

### First Time Setup

When you run `vibecontainer create` for the first time:

1. The TUI wizard prompts you for required credentials
2. After successful stack creation, credentials are automatically saved to your system keychain
3. On macOS: stored in Keychain Access
4. On Windows: stored in Credential Manager
5. On Linux: stored in Secret Service (GNOME Keyring, KWallet, etc.)

### Subsequent Use

On future runs of `vibecontainer create`:

1. The CLI automatically loads stored credentials from the keychain
2. Password fields show a hint: "Using saved credential (press Enter to keep, or type new value)"
3. Press Enter to use the saved credential
4. Or type a new value to update it
5. New/updated credentials are automatically saved

### Command-Line Flags

Command-line flags always take precedence over stored credentials:

```sh
# This will use the provided token instead of any stored credential
vibecontainer create --claude-oauth-token "new-token-here"
```

### Opt-Out

To prevent credentials from being saved:

```sh
vibecontainer create --no-save-auth
```

## Managing Credentials

### List Stored Credentials

View which credentials are currently stored (without showing actual values):

```sh
vibecontainer credentials list
```

Example output:
```
Stored credentials:
  Claude OAuth Token:      ✓ stored
  Anthropic API Key:       ✗ not stored
  Codex Auth JSON:         ✗ not stored
  OpenAI API Key:          ✓ stored
  Codex API Key:           ✗ not stored
  Tunnel Token:            ✓ stored
```

### Clear All Credentials

Remove all stored credentials from the keychain:

```sh
vibecontainer credentials clear
```

This is useful when:
- Switching accounts
- Revoking access
- Troubleshooting authentication issues
- Preparing to hand off a machine

## Security

- Credentials are stored using native OS keychain APIs
- Never stored in plain text files
- Protected by your OS user account permissions
- On macOS: can be viewed/managed in Keychain Access app
- On Windows: can be viewed/managed in Credential Manager
- On Linux: protected by your keyring password

## Cross-Platform Support

The credential storage uses [go-keyring](https://github.com/zalando/go-keyring), which provides unified access to:

| Platform | Backend |
|----------|---------|
| macOS | Keychain |
| Windows | Credential Manager |
| Linux | Secret Service (GNOME Keyring, KWallet) |
| Other Unix | Fallback to encrypted file |

## Troubleshooting

### "Failed to save credentials to keychain"

This warning appears if the CLI can't access your system keychain. Possible causes:

- **Linux**: Secret Service daemon not running
  - Install `gnome-keyring` or `kwalletmanager`
  - Start the service: `gnome-keyring-daemon --start`
- **macOS**: Keychain Access permissions issue
  - Open Keychain Access and verify you can unlock it
- **Headless/SSH**: No keyring service available
  - Use `--no-save-auth` flag to skip saving

### Clear credentials if you see authentication errors

```sh
vibecontainer credentials clear
vibecontainer create  # Will prompt for fresh credentials
```

### Multiple accounts/profiles

The current implementation stores one set of credentials per type. To use multiple accounts:

1. Clear credentials before switching: `vibecontainer credentials clear`
2. Or use command-line flags to override: `--claude-oauth-token`
3. Or use `--no-save-auth` to avoid overwriting stored credentials

## Examples

### Claude with OAuth (stored credentials)

```sh
# First time - enter credentials in TUI
vibecontainer create --provider claude

# Next time - credentials auto-loaded
vibecontainer create --provider claude  # No need to re-enter!
```

### Codex with API key (command-line)

```sh
# Provide via flag (will be saved for next time)
vibecontainer create --provider codex --openai-api-key "sk-..."

# Next time - no flag needed
vibecontainer create --provider codex
```

### Temporary stack (don't save credentials)

```sh
# Use --no-save-auth for one-off testing
vibecontainer create --no-save-auth \
  --provider claude \
  --claude-oauth-token "temporary-token"
```

## Implementation Details

- Service name in keychain: `vibecontainer`
- Key names follow pattern: `claude_oauth_token`, `anthropic_api_key`, etc.
- Credentials are stored individually (not as a bundle)
- Empty/blank credentials are not saved to avoid cluttering the keychain
- Load failures are silent (returns empty string)
- Save failures log a warning but don't stop stack creation
