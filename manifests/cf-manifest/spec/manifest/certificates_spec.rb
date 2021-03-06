RSpec.describe "certificates" do
  def get_all_cas_usages(o)
    return o.flat_map { |v| get_all_cas_usages(v) } if o.is_a? Array

    if o.is_a? Hash
      # Match checks if it is a usage ("name.value") or a variable ("name")
      return [o['ca']] if o['ca']&.match?(/[.]/)
      return [o['ca_cert']] if o['ca_cert']&.match?(/[.]/)

      return o.values.flat_map { |v| get_all_cas_usages(v) }
    end

    []
  end

  context "ca certificates" do
    let(:manifest) { manifest_with_defaults }

    let(:ca_usages) do
      get_all_cas_usages(manifest.fetch('.')).map do |usage|
        usage.gsub(/[()]/, '') # delete surrounding parens
      end
    end

    it "should detect some ca certificate usages" do
      expect(ca_usages).not_to eq([])
    end

    it "should use .ca for every usage of a ca certificate" do
      ca_usages.each do |ca_usage|
        expect(ca_usage).to match(/[.]ca$/),
          "Usage of CA #{ca_usage} should be cert_name.ca not ca_name.certificate, otherwise credhub rotation will fail"
      end
    end
  end
end
