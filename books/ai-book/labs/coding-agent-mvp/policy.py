from models import ToolCall


READ_ONLY_TOOLS = {"list_files", "read_file", "search_code", "git_diff"}
WRITE_TOOLS = {"replace_in_file", "create_file"}
SHELL_TOOLS = {"run_shell"}


class Policy:
    def __init__(self, auto_edit: bool = True, auto_shell: bool = False):
        self.auto_edit = auto_edit
        self.auto_shell = auto_shell

    def decide(self, call: ToolCall) -> str:
        if call.name in READ_ONLY_TOOLS:
            return "allow"
        if call.name in WRITE_TOOLS:
            return "allow" if self.auto_edit else "ask"
        if call.name in SHELL_TOOLS:
            return "allow" if self.auto_shell else "ask"
        return "deny"
