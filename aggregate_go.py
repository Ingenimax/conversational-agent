import os

def aggregate_go_files_with_file_headers(source_dir, output_file):
    """
    Aggregates all *.go files from the source directory and its subfolders
    into a single file, with a comment indicating the path of each file.

    :param source_dir: Directory to scan for .go files
    :param output_file: File where the aggregated content will be saved
    """
    package_declared = False
    imports_collected = set()
    go_code_lines = []

    for root, _, files in os.walk(source_dir):
        for file in files:
            if file.endswith('.go'):
                file_path = os.path.join(root, file)
                with open(file_path, 'r') as f:
                    # Add file header
                    relative_path = os.path.relpath(file_path, source_dir)
                    go_code_lines.append(f"\n# {relative_path}\n")

                    in_import_block = False
                    for line in f:
                        stripped_line = line.strip()

                        # Skip package declaration except for the first file
                        if stripped_line.startswith('package '):
                            if not package_declared:
                                go_code_lines.append(line)
                                package_declared = True
                            continue

                        # Handle imports
                        if stripped_line.startswith('import '):
                            if stripped_line == 'import':
                                in_import_block = True
                                continue
                            if stripped_line not in imports_collected:
                                imports_collected.add(stripped_line)
                                go_code_lines.append(line)
                            continue

                        if in_import_block:
                            if stripped_line == ')':
                                in_import_block = False
                            elif stripped_line not in imports_collected:
                                imports_collected.add(stripped_line)
                                go_code_lines.append(line)
                            continue

                        # Add other lines
                        go_code_lines.append(line)

    # Write the aggregated content to the output file
    with open(output_file, 'w') as out_file:
        out_file.writelines(go_code_lines)

    print(f"All .go files aggregated into {output_file}")


# Example Usage
source_directory = "."  # Replace with your source directory path
output_filename = "full-project.txt"
aggregate_go_files_with_file_headers(source_directory, output_filename)