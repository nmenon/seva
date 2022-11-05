// This is broken out in case extra metdata parsing is used

enum UrlType {
  compose,
  metadata,
}

String build_url(String app_name, UrlType url_type) {
  // build the requested url with the given app name
  String store_url =
      "https://raw.githubusercontent.com/StaticRocket/seva-apps/main/";
  store_url += "${app_name}/";
  if (url_type == UrlType.compose)
    store_url += "docker-compose.yml";
  else
    store_url += "metadata.json";
  return store_url;
}

class AppMetadata {
  // class to hold application data
   String name = "No app selected";
   String note = "Select an app from the store";
   String source_url = "";
   bool   has_web_interface = false;

  AppMetadata(this.name, this.note, this.source_url, this.has_web_interface);

  AppMetadata.empty();

  AppMetadata.from_json(Map<String, dynamic> json) {
    if (json['name']) {
      name = json['name'];
      note = json['note'];
      source_url = json['source_url'];
      has_web_interface = json['has_web_interface'];
    }
  }
}
