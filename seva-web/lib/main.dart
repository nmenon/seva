import 'package:http/http.dart' as http;
import 'package:flutter/material.dart';
import 'dart:html' as html;
import 'dart:io';
import 'dart:convert';
import 'url_builder.dart';
import 'websocket.dart';
import 'navigation_menu.dart';

// store url, must point to page with proper message listener
final String store_url = 'http://${host_name}:8001/';

// Global key in case we want to use more snackbar messages
final GlobalKey<ScaffoldMessengerState> rootScaffoldMessengerKey =
    GlobalKey<ScaffoldMessengerState>();

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({Key? key}) : super(key: key);
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'seva',
      theme: ThemeData(
        primarySwatch: Colors.red,
      ),
      home: const MyHomePage(title: 'Seva Control Center'),
      scaffoldMessengerKey: rootScaffoldMessengerKey,
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({Key? key, required this.title}) : super(key: key);
  final String title;

  @override
  State<MyHomePage> createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  final GlobalKey<WebSocketStatusState> _websocket_key = GlobalKey();
  AppMetadata _selected_app = AppMetadata.empty();
  String _message_data = '';
  bool _store_connected = false;

  void show_warning(String message) {
    // use the scaffold key to display a message at the root of the app
    rootScaffoldMessengerKey.currentState
        ?.showSnackBar(SnackBar(content: Text(message)));
  }

  void _app_hook(event) async {
    // store hook handler
    if (event.data == 'seva-init')
      setState(() {
        _store_connected = true;
      });
    else {
      _message_data = event.data;
      await _fetch_app_metadata();
    }
  }

  void _launch_app_browser() async {
    // launch the store at the given url and perform initial handshake
    var popup = html.window.open(store_url, "blank_");
    html.window.addEventListener("message", _app_hook, false);
    while (_store_connected == false) {
      await Future.delayed(Duration(seconds: 1));
      popup.postMessage("seva-init", store_url);
    }
  }

  Future<void> _fetch_app_metadata() async {
    // fetch and test app metadata
    String metadata_url = build_url(_message_data, UrlType.metadata);
    final http.Response response = await http.get(Uri.parse(metadata_url));
    if (response.statusCode == 200) {
      var recieved_json = jsonDecode(response.body);
      setState(() {
        _selected_app = AppMetadata.from_json(recieved_json);
      });
      await _websocket_key.currentState?.load_app(_message_data);
    } else {
      show_warning('Failed to load data');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
        drawer: NavigationMenu(),
        appBar: AppBar(
          title: Text(widget.title),
        ),
        body: Center(
          child: WebSocketStatus(key: _websocket_key),
        ),
        floatingActionButton: Column(
          mainAxisAlignment: MainAxisAlignment.end,
          children: <Widget>[
            Padding(
              padding: EdgeInsets.all(16),
              child: _store_connected
                  ? Tooltip(
                      message: "Store handshake has occurred successfully",
                      child: const Icon(Icons.sync),
                    )
                  : Tooltip(
                      message: "Store handshake has not occurred",
                      child: const Icon(Icons.sync_disabled),
                    ),
            ),
            FloatingActionButton(
              onPressed: _launch_app_browser,
              tooltip: 'Store',
              child: const Icon(Icons.apps),
            ),
          ],
        ));
  }
}
