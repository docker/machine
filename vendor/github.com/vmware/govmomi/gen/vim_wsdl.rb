# Copyright (c) 2014 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

require "nokogiri"
require "test/unit"

class Peek
  class Type
    attr_accessor :parent, :children, :klass

    def initialize(name)
      @name = name
      @children = []
    end

    def base?
      return !children.empty?
    end
  end

  @@types = {}
  @@refs = {}
  @@enums = {}

  def self.types
    return @@types
  end

  def self.refs
    return @@refs
  end

  def self.enums
    return @@enums
  end

  def self.ref(type)
    refs[type] = true
  end

  def self.enum(type)
    enums[type] = true
  end

  def self.enum?(type)
    enums[type]
  end

  def self.register(name)
    raise unless name
    types[name] ||= Type.new(name)
  end

  def self.base?(name)
    return unless c = types[name]
    c.base?
  end

  def self.dump_interfaces(io)
    types.keys.sort.each do |name|
      next unless base?(name)

      types[name].klass.dump_interface(io, name)
    end
  end
end

class EnumValue
  def initialize(type, value)
    @type = type
    @value = value
  end

  def type_name
    @type.name
  end

  def var_name
    n = @type.name
    v = var_value
    if v == ""
      n += "Null"
    else
      n += (v[0].capitalize + v[1..-1])
    end

    return n
  end

  def var_value
    @value
  end

  def dump(io)
    io.print "%s = %s(\"%s\")\n" % [var_name, type_name, var_value]
  end
end

class Simple
  include Test::Unit::Assertions

  attr_accessor :name, :type

  def initialize(node)
    @node = node
  end

  def name
    @name || @node["name"]
  end

  def type
    @type || @node["type"]
  end

  def is_enum?
    false
  end

  def dump_init(io)
    # noop
  end

  def var_name
    n = self.name
    n = n[1..-1] if n[0] == "_" # Strip leading _
    n = n[0].capitalize + n[1..-1] # Capitalize
    return n
  end

  def vim_type?
    ns, _ = self.type.split(":", 2)
    ns == "vim25"
  end

  def vim_type(t = self.type)
    ns, t = t.split(":", 2)
    raise if ns != "vim25"
    t
  end

  def base_type?
    vim_type? && Peek.base?(vim_type)
  end

  def enum_type?
    vim_type? && Peek.enum?(vim_type)
  end

  def any_type?
    self.type == "xsd:anyType"
  end

  def var_type
    t = self.type
    prefix = ""

    prefix += "[]" if slice?

    if t =~ /^xsd:(.*)$/
      t = $1
      case t
      when "string"
      when "int"
      when "boolean"
        t = "bool"
        if !slice? && optional?
          prefix += "*"
          self.need_omitempty = false
        end
      when "long"
        t = "int64"
      when "dateTime"
        t = "time.Time"
        if !slice? && optional?
          prefix += "*"
          self.need_omitempty = false
        end
      when "anyType"
        t = "AnyType"
      when "byte"
      when "double"
        t = "float64"
      when "float"
        t = "float32"
      when "short"
        t = "int16"
      when "base64Binary"
        t = "[]byte"
      when "anyURI"
        t = "url.URL"
      else
        raise "unknown type: %s" % t
      end
    else
      t = vim_type
      if base_type?
        prefix += "Base"
      else
        prefix += "*" if !slice? && !enum_type? && optional?
      end
    end

    prefix + t
  end

  def slice?
    test_attr("maxOccurs", "unbounded")
  end

  def optional?
    test_attr("minOccurs", "0")
  end

  def need_omitempty=(v)
    @need_omitempty = v
  end

  def need_omitempty?
    var_type # HACK: trigger setting need_omitempty if necessary
    if @need_omitempty.nil?
      @need_omitempty = optional?
    else
      @need_omitempty
    end
  end

  def need_typeattr?
    base_type? || any_type?
  end

  protected

  def test_attr(attr, expected)
    actual = @node.attr(attr)
    if actual != nil
      case actual
      when expected
        true
      else
        raise "%s=%s" % [value, type.attr(value)]
      end
    else
      false
    end
  end
end

class Element < Simple
  def initialize(node)
    super(node)
  end

  def has_type?
    !@node["type"].nil?
  end

  def child
    cs = @node.element_children
    assert_equal 1, cs.length
    assert_equal "complexType", cs.first.name

    t = ComplexType.new(cs.first)
    t.name = self.name
    t
  end

  def dump(io)
    if has_type?
      io.print "type %s %s\n\n" % [name, var_type]
    else
      child.dump(io)
    end
  end

  def dump_init(io)
    if has_type?
      io.print "func init() {\n"
      io.print "t[\"%s\"] = reflect.TypeOf((*%s)(nil)).Elem()\n" % [name, name]
      io.print "}\n\n"
    end
  end

  def dump_field(io)
    tag = name
    tag += ",omitempty" if need_omitempty?
    tag += ",typeattr" if need_typeattr?
    io.print "%s %s `xml:\"%s\"`\n" % [var_name, var_type, tag]
  end

  def peek(type=nil)
    if has_type?
      return if self.type =~ /^xsd:/

      Peek.ref(vim_type)
    else
      child.peek()
    end
  end
end

class Attribute < Simple
  def dump_field(io)
    tag = name
    tag += ",omitempty" if need_omitempty?
    tag += ",attr"
    io.print "%s %s `xml:\"%s\"`\n" % [var_name, var_type, tag]
  end
end

class SimpleType < Simple
  def is_enum?
    true
  end

  def dump(io)
    enums = @node.xpath(".//xsd:enumeration").map do |n|
      EnumValue.new(self, n["value"])
    end

    io.print "type %s string\n\n" % name
    io.print "const (\n"
    enums.each { |e| e.dump(io) }
    io.print ")\n\n"
  end

  def dump_init(io)
    io.print "func init() {\n"
    io.print "t[\"%s\"] = reflect.TypeOf((*%s)(nil)).Elem()\n" % [name, name]
    io.print "}\n\n"
  end

  def peek
    Peek.enum(name)
  end
end

class ComplexType < Simple
  class SimpleContent < Simple
    def dump(io)
      attr = Attribute.new(@node.at_xpath(".//xsd:attribute"))
      attr.dump_field(io)

      # HACK DELUXE(PN)
      extension = @node.at_xpath(".//xsd:extension")
      type = extension["base"].split(":", 2)[1]
      io.print "Value %s `xml:\",chardata\"`\n" % type
    end

    def peek
    end
  end

  class ComplexContent < Simple
    def base
      extension = @node.at_xpath(".//xsd:extension")
      assert_not_nil extension

      base = extension["base"]
      assert_not_nil base

      vim_type(base)
    end

    def dump(io)
      Sequence.new(@node).dump(io, base)
    end

    def dump_interface(io, name)
      Sequence.new(@node).dump_interface(io, name)
    end

    def peek
      Sequence.new(@node).peek(base)
    end
  end

  class Sequence < Simple
    def sequence
      sequence = @node.at_xpath(".//xsd:sequence")
      if sequence != nil
        sequence.element_children.map do |n|
          Element.new(n)
        end
      else
        nil
      end
    end

    def dump(io, base = nil)
      return unless elements = sequence

      io.print "%s\n\n" % base

      elements.each do |e|
        e.dump_field(io)
      end
    end

    def dump_interface(io, name)
      method = "Get%s() *%s" % [name, name]
      io.print "func (b *%s) %s { return b }\n" % [name, method]
      io.print "type Base%s interface {\n" % name
      io.print "%s\n" % method
      io.print "}\n\n"
      io.print "func init() {\n"
      io.print "t[\"Base%s\"] = reflect.TypeOf((*%s)(nil)).Elem()\n" % [name, name]
      io.print "}\n\n"
    end

    def peek(base = nil)
      return unless elements = sequence
      name = @node.attr("name")
      return unless name

      elements.each do |e|
        e.peek(name)
      end

      c = Peek.register(name)
      if base
        c.parent = base
        Peek.register(c.parent).children << name
      end
    end
  end

  def klass
    @klass ||= begin
                 cs = @node.element_children
                 if !cs.empty?
                   assert_equal 1, cs.length

                   case cs.first.name
                   when "simpleContent"
                     SimpleContent.new(@node)
                   when "complexContent"
                     ComplexContent.new(@node)
                   when "sequence"
                     Sequence.new(@node)
                   else
                     raise "don't know what to do for element: %s..." % cs.first.name
                   end
                 end
               end
  end

  def dump_init(io)
    io.print "func init() {\n"
    io.print "t[\"%s\"] = reflect.TypeOf((*%s)(nil)).Elem()\n" % [name, name]
    io.print "}\n\n"
  end

  def dump(io)
    io.print "type %s struct {\n" % name
    klass.dump(io) if klass
    io.print "}\n\n"
  end

  def peek
    Peek.register(name).klass = klass
    klass.peek if klass
  end
end

class Schema
  include Test::Unit::Assertions

  def initialize(xml, parent = nil)
    @xml = Nokogiri::XML.parse(xml)
  end

  def targetNamespace
    @xml.root["targetNamespace"]
  end

  # We have some assumptions about structure, make sure they hold.
  def validate_assumptions!
    # Every enumeration is part of a restriction
    @xml.xpath(".//xsd:enumeration").each do |n|
      assert_equal "restriction", n.parent.name
    end

    # See type == enum
    @xml.xpath(".//xsd:restriction").each do |n|
      # Every restriction has type xsd:string (it's an enum)
      assert_equal "xsd:string", n["base"]

      # Every restriction is part of a simpleType
      assert_equal "simpleType", n.parent.name

      # Every restriction is alone
      assert_equal 1, n.parent.element_children.size
    end

    # See type == complex_content
    @xml.xpath(".//xsd:complexContent").each do |n|
      # complexContent is child of complexType
      assert_equal "complexType", n.parent.name

    end

    # See type == complex_type
    @xml.xpath(".//xsd:complexType").each do |n|
      cc = n.element_children

      # OK to have an empty complexType
      next if cc.size == 0

      # Require 1 element otherwise
      assert_equal 1, cc.size

      case cc.first.name
      when "complexContent"
        # complexContent has 1 "extension" element
        cc = cc.first.element_children
        assert_equal 1, cc.size
        assert_equal "extension", cc.first.name

        # extension has 1 "sequence" element
        ec = cc.first.element_children
        assert_equal 1, ec.size
        assert_equal "sequence", ec.first.name

        # sequence has N "element" elements
        sc = ec.first.element_children
        assert sc.all? { |e| e.name == "element" }
      when "simpleContent"
        # simpleContent has 1 "extension" element
        cc = cc.first.element_children
        assert_equal 1, cc.size
        assert_equal "extension", cc.first.name

        # extension has 1 or more "attribute" elements
        ec = cc.first.element_children
        assert_not_equal 0, ec.size
        assert_equal "attribute", ec.first.name
      when "sequence"
        # sequence has N "element" elements
        sc = cc.first.element_children
        assert sc.all? { |e| e.name == "element" }
      else
        raise "unknown element: %s" % cc.first.name
      end
    end

    includes.each do |i|
      i.validate_assumptions!
    end
  end

  def types
    return to_enum(:types) unless block_given?

    includes.each do |i|
      i.types do |t|
        yield t
      end
    end

    @xml.root.children.each do |n|
      case n.class.to_s
      when "Nokogiri::XML::Text"
        next
      when "Nokogiri::XML::Element"
        case n.name
        when "include", "import"
          next
        when "element"
          yield Element.new(n)
        when "simpleType"
          yield SimpleType.new(n)
        when "complexType"
          yield ComplexType.new(n)
        else
          raise "unknown child: %s" % n.name
        end
      else
        raise "unknown type: %s" % n.class
      end
    end
  end

  def includes
    @includes ||= @xml.root.xpath(".//xmlns:include").map do |n|
      Schema.new(WSDL.read n["schemaLocation"])
    end
  end
end


class Operation
  include Test::Unit::Assertions

  def initialize(wsdl, operation_node)
    @wsdl = wsdl
    @operation_node = operation_node
  end

  def name
    @operation_node["name"]
  end

  def remove_ns(x)
    ns, x = x.split(":", 2)
    assert_equal "vim25", ns
    x
  end

  def find_type_for(type)
    type = remove_ns(type)

    message = @wsdl.message(type)
    assert_not_nil message

    part = message.at_xpath("./xmlns:part")
    assert_not_nil message

    remove_ns(part["element"])
  end

  def input
    type = @operation_node.at_xpath("./xmlns:input").attr("message")
    find_type_for(type)
  end

  def go_input
    "types." + input
  end

  def output
    type = @operation_node.at_xpath("./xmlns:output").attr("message")
    find_type_for(type)
  end

  def go_output
    "types." + output
  end

  def dump(io)
    io.print <<EOS
  type #{name}Body struct{
    Req *#{go_input} `xml:"urn:vim25 #{input},omitempty"`
    Res *#{go_output} `xml:"urn:vim25 #{output},omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *#{name}Body) Fault() *soap.Fault { return b.Fault_ }

EOS

    io.print "func %s(ctx context.Context, r soap.RoundTripper, req *%s) (*%s, error) {\n" % [name, go_input, go_output]
    io.print <<EOS
  var reqBody, resBody #{name}Body

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
EOS

    io.print "}\n\n"
  end
end

class WSDL
  attr_reader :xml

  PATH = File.expand_path("../sdk", __FILE__)

  def self.read(file)
    File.open(File.join(PATH, file))
  end

  def initialize(xml)
    @xml = Nokogiri::XML.parse(xml)
  end

  def validate_assumptions!
    schemas.each do |s|
      s.validate_assumptions!
    end
  end

  def types(&blk)
    return to_enum(:types) unless block_given?

    schemas.each do |s|
      s.types(&blk)
    end
  end

  def schemas
    @schemas ||= @xml.xpath('.//xmlns:types/xsd:schema').map do |n|
      Schema.new(n.to_xml)
    end
  end

  def operations
    @operations ||= @xml.xpath('.//xmlns:portType/xmlns:operation').map do |o|
      Operation.new(self, o)
    end
  end

  def message(type)
    @messages ||= begin
                    h = {}
                    @xml.xpath('.//xmlns:message').each do |n|
                      h[n.attr("name")] = n
                    end
                    h
                  end

    @messages[type]
  end

  def peek
    types.
      sort_by { |x| x.name }.
      uniq { |x| x.name }.
      select { |x| x.name[0] == x.name[0].upcase }. # Only capitalized methods for now...
      each { |e| e.peek() }
  end

  def self.header(name)
    return <<EOF
/*
Copyright (c) 2014 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package #{name}

EOF
  end
end
