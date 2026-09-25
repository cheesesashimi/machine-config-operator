function walk(node, visitor) {
  visitor(node);
  if (node.children && Array.isArray(node.children)) {
    node.children.forEach((child) => walk(child, visitor));
  }
}

export default function remarkStripWordProcessorArtifacts() {
  return (tree) => {
    walk(tree, (node) => {
      // Frontmatter and code are data, so only normalize prose text nodes.
      if (node.type === 'text' && typeof node.value === 'string') {
        node.value = node.value
          .replace(/[“”]/g, '"')
          .replace(/[‘’]/g, "'")
          .replace(/–/g, '-')
          .replace(/—/g, '--')
          .replace(/…/g, '...')
          .replace(/\u00a0/g, ' ');
      }
    });
  };
}
