require 'asciidoctor/extensions' unless RUBY_ENGINE == 'opal'
require 'rouge'

include Asciidoctor

class Proto2asciidocInlineMacro < Extensions::InlineMacroProcessor
  use_dsl

  named :proto2asciidoc
  name_positional_attributes 'key', 'value'

  def process parent, target, attrs
    if target == 'tag'
      tags = attrs['key'].split('%')
      o = ''
      tags.each do |set|
        set.gsub!("any","&#42;")
        t = set.split('@')
        if parent.document.basebackend? 'html'
          o = o + sprintf('<span class="tag"><span class="key">%s</span><span class="data">%s</span></span>', t[0], t[1])
        else
          if set.equal?(tags.last)
            o = o + sprintf('%s=%s', t[0], t[1])
          else
            o = o + sprintf('%s=%s,', t[0], t[1])
          end
        end
      end
      if parent.document.basebackend? 'html'
        out = sprintf('<var class="proto2asciidoc tags"><span class="fa icon"></span><span class="contain">%s</span></var>', o)
      else
        out = sprintf('<code>%s</code>', o)
      end
      %(#{out})
    elsif target == 'endpoint' || target == 'enum' || target == 'message'
        text = target
        key = attrs['key']
        value = attrs['value']
      if !attrs.key?('value')
        link = sprintf('#%s', key)
        if key.include? "proto."
          link = sprintf('#%s', key.delete_prefix('proto.'))
        end
        if target == 'message'
          link = sprintf('%s_message', link)
        elsif target == 'enum'
          link = sprintf('%s_enum', link)
        end
        text = '<span class="fa icon"></span><span class="text">'+ key + '</span>'
      else
        if key.include? "proto."
          key = key.delete_prefix('proto.')
        end
        text = '<span class="fa icon"></span><span class="text">'+ key + ' ' + value + '</span>'
        link = sprintf('#%s_%s', key, value)
      end
      
      %(<var class="proto2asciidoc #{target}">#{(create_anchor parent, text, type: :link, target: link.downcase).render}</var>)
    else
      key = attrs['key']
      link = sprintf('#%s', key)
      if target == "command"
        link = sprintf('#%s_command', key)
      elsif target == "service"
        link = sprintf('#%s_service', key)
      end
      text = target
      text = '<span class="fa icon"></span><span class="text">'+ key + '</span>'
      %(<var class="proto2asciidoc #{target}">#{(create_anchor parent, text, type: :link, target: link.downcase).render}</var>)
    end
  end
end