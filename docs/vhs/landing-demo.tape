# Generate:
# vhs < docs/vhs/landing-demo.tape
# Requires ffmeg and aqua install
# brew install ffmpeg ttyd

# Where should we write the GIF?
Output docs/vhs/assets/landing-demo.gif

# Set up a 1200x600 terminal with 46px font.
Set FontSize 18
Set Width 1200
Set Height 800

Type "dsv --help"
Enter
Sleep 2s
Type "dsv secret read --help"
Enter
Sleep 2s
Type "dsv secret read --path 'ci-secrets:tokens:mytoken' --filter '.data.token' --plain"
Sleep 2s
