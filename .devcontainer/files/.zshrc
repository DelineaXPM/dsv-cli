#eval "$(direnv hook zsh)"
export ZSH="$HOME/.oh-my-zsh"
ZSH_THEME=avit
plugins=(z direnv zsh-interactive-cd docker golang gh zsh-navigation-tools)
source "$ZSH/oh-my-zsh.sh"



# Display optional first run image specific notice if configured and terminal is interactive
# if [ -t 1 ] && [[ "${TERM_PROGRAM}" = "vscode" || "${TERM_PROGRAM}" = "codespaces" ]] && [ ! -f "$HOME/.config/vscode-dev-containers/first-run-notice-already-displayed" ]; then
if [ -t 1 ] && [[ "${TERM_PROGRAM}" = "vscode" || "${TERM_PROGRAM}" = "codespaces" ]]; then
    if [ -f "$HOME/first-run-notice.txt" ]; then
        if  command -v glow &>/dev/null;   then
            glow "$HOME/first-run-notice.txt"
        else
            cat "$HOME/first-run-notice.txt"
        fi
    fi
    mkdir -p "$HOME/.config/vscode-dev-containers" || echo 'not able to create "$HOME/.config/vscode-dev-containers"'
    # Mark first run notice as displayed after 10s to avoid problems with fast terminal refreshes hiding it
    ((sleep 10s; touch "$HOME/.config/vscode-dev-containers/first-run-notice-already-displayed") &)
fi