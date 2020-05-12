RUBY_ENGINE == 'opal' ? (require 'proto2asciidoc-inline-macro/extension') : (require_relative 'proto2asciidoc-inline-macro/extension')

Asciidoctor::Extensions.register do
  if @document.basebackend? 'html'
    inline_macro Proto2asciidocInlineMacro
  elsif @document.basebackend? 'docbook'
    inline_macro Proto2asciidocInlineMacro
  end
end