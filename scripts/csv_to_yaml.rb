#
# This is a really fast and dirty script that I used to migrate photos out of
# Black Swan and Flickr and over to a Dropbox solution. I'm keeping it in the
# source tree for posterity and an example of using the Dropbox API even though
# I don't really expect to use it again.
#
# Postgres invocations to get a CSV:
#
#     CREATE TEMP TABLE photographs (id BIGINT, occurred_at TIMESTAMPTZ, title TEXT, description TEXT);
#
#     INSERT INTO photographs (SELECT slug::BIGINT, occurred_at, metadata -> 'title', metadata -> 'description' FROM events WHERE type = 'flickr' ORDER BY occurred_at);
#
#     \COPY photographs TO './photographs.csv'
#
#

require 'csv'
require 'excon'
require 'json'
require 'yaml'

# Brandur Personal App
API_KEY = "..."

csv_file = './photographs.csv'
yaml_file = './photographs.yaml'

data = []

def get_link(slug)
  dropbox_path = "/photo-archive/_lifestream/#{slug}.jpg"
  p dropbox_path

  res = Excon.post("https://api.dropboxapi.com/2/sharing/list_shared_links",
    headers: {
      "Authorization" => "Bearer #{API_KEY}",
      "Content-Type" => "application/json",
    },
    body: JSON.generate({
      path: dropbox_path,
      direct_only: true,
    }),
    expects: [200, 201]
  )
  data = JSON.parse(res.body)

  link = data["links"].first

  unless link
    res = Excon.post("https://api.dropboxapi.com/2/sharing/create_shared_link_with_settings",
      headers: {
        "Authorization" => "Bearer #{API_KEY}",
        "Content-Type" => "application/json",
      },
      body: JSON.generate({
        path: dropbox_path,
        settings: {
          requested_visibility: "public",
        },
      }),
      expects: [200, 201]
    )
    link = JSON.parse(res.body)
  end

  url = link["url"].gsub(/\?dl=0/, "?dl=1")
  p url
  url
end

begin
  CSV.foreach(csv_file, { col_sep: "\t", quote_char: "|" }) do |row|

    p row
    data << {
      slug: row[0],
      occurred_at: DateTime.parse(row[1]).rfc3339,
      title: row[2] || "",
      description: row[3] || "",
      original_image: get_link(row[0]),
    }

    # hopefully avoid rate limiting
    sleep(0.2)
  end

  wrapper = {photographs: data}

  doc_yaml = YAML.dump(wrapper)
  File.open(yaml_file, 'w') { |file| file.write(doc_yaml) }

rescue Excon::Error::BadRequest, Excon::Error::Conflict => e
  abort("#{e}: #{e.response.body}")
end
