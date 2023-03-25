import 'package:flutter/material.dart';
import 'websocket.dart';
import 'dart:async';
import 'dart:html';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';

class WebProxy extends StatefulWidget {
  @override
  State<WebProxy> createState() => _WebProxyState();
}

class _WebProxyState extends State<WebProxy> {
  var proxy_http = TextEditingController();
  var proxy_no = TextEditingController();

  final _form = GlobalKey<FormState>();

  bool waiting_on_response_ = false;

  Future<WebSocketCommand> response_handler() async {
    // catch the response code and update state accordingly
    setState(() {
      waiting_on_response_ = true;
    });
    String response = await stream.first;
    setState(() {
      waiting_on_response_ = false;
    });
    return WebSocketCommand.from_json(jsonDecode(response));
  }

  Future<void> save_settings(var serialized_settings) async {
    WebSocketCommand.outbound("save_settings", [serialized_settings]).send();
    WebSocketCommand command = await response_handler();
    if (command.response[0] == '1') {
      // TODO: Error handling
    }
  }

  bool isValidUrl(var proxy_url) {
    var urlPattern =
        r"(http|https|socks4|socks5)://[A-Za-z0-9\-._~:/?#\[\]@!$&'\(\)*+,;%=]+";
    var regExp = new RegExp(urlPattern);
    return regExp.hasMatch(proxy_url);
  }

  Widget proxy_settings(BuildContext context) {
    return Container(
      padding: EdgeInsets.all(8.0),
      child: Column(
        children: <Widget>[
          Text(
            'Proxy Settings',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          Container(
            padding: EdgeInsets.all(8.0),
            child: TextFormField(
              controller: proxy_http,
              keyboardType: TextInputType.url,
              decoration: const InputDecoration(
                labelText:
                    "Enter a HTTP/SOCKS URL or IP to proxy traffic through",
              ),
              validator: (value) {
                if (value == null ||
                    value.isEmpty ||
                    isValidUrl(proxy_http.text)) {
                  return null;
                }
                return "Please enter a valid URL or IP";
              },
            ),
          ),
          Container(
            padding: EdgeInsets.all(8.0),
            child: TextFormField(
              controller: proxy_no,
              keyboardType: TextInputType.url,
              decoration: const InputDecoration(
                labelText:
                    "Enter URLs / IPs that should not be routed through the proxy",
              ),
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
        appBar: AppBar(
          title: const Text("Settings"),
        ),
        body: Form(
          key: _form, //assigning key to form
          child: Column(children: [proxy_settings(context)]),
        ),
        floatingActionButton: Column(
          mainAxisAlignment: MainAxisAlignment.end,
          children: <Widget>[
            FloatingActionButton(
              onPressed: () {
                if (_form.currentState!.validate()) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Applying settings...')),
                  );

                  var settings = {
                    "https_proxy": proxy_http.text,
                    "http_proxy": proxy_http.text,
                    "ftp_proxy": proxy_http.text,
                    "no_proxy": proxy_no.text
                  };
                  var serialized_settings = json.encode(settings);
                  save_settings(serialized_settings);
                }
              },
              tooltip: 'Save',
              child: const Icon(Icons.save),
            ),
          ],
        ));
  }
}
