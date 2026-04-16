import sys
import re

def analyze_tags(filepath):
    with open(filepath, 'r') as f:
        content = f.read()
    
    # Remove strings and comments to avoid false positives
    content = re.sub(r'{/\*.*?\*/}', '', content, flags=re.DOTALL)
    content = re.sub(r'//.*?\n', '\n', content)
    content = re.sub(r'`.*?`', '""', content, flags=re.DOTALL)
    content = re.sub(r'".*?"', '""', content)
    content = re.sub(r"'.*?'", "''", content)

    stack = []
    lines = content.split('\n')
    
    for i, line in enumerate(lines):
        line_num = i + 1
        # Find all divs and portals
        tokens = re.findall(r'<(div|Portal|/div|/Portal)[^>]*>', line)
        for token in tokens:
            if token == 'div' or token.startswith('div '):
                if not token.endswith('/'):
                    stack.append(('div', line_num))
            elif token == '/div':
                if stack and stack[-1][0] == 'div':
                    stack.pop()
                else:
                    print(f"Extra </div> at line {line_num}")
            elif token == 'Portal' or token.startswith('Portal '):
                if not token.endswith('/'):
                    stack.append(('Portal', line_num))
            elif token == '/Portal':
                if stack and stack[-1][0] == 'Portal':
                    stack.pop()
                else:
                    print(f"Extra </Portal> at line {line_num} (Expected {stack[-1] if stack else 'Nothing'})")
                    if stack: stack.pop() # Force pop to continue

    print("\nSummary of unclosed tags:")
    for tag, line in stack:
        print(f"Unclosed <{tag}> from line {line}")

analyze_tags(sys.argv[1])
