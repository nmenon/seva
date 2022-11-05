import 'dart:convert';
import 'package:flutter/material.dart';
import 'dart:html' as html;
import 'package:web_socket_channel/web_socket_channel.dart';
import 'url_builder.dart';

// setup the websocket
final host_name = html.window.location.hostname;
var channel = WebSocketChannel.connect(
  Uri.parse('ws://${host_name}:8000/ws'),
);
Stream stream = channel.stream.asBroadcastStream();

class WebSocketStatus extends StatefulWidget {
  const WebSocketStatus({Key? key}) : super(key: key);
  @override
  State<WebSocketStatus> createState() => WebSocketStatusState();
}

class WebSocketStatusState extends State<WebSocketStatus> {
  AppMetadata _selected_app = AppMetadata.empty();
  bool waiting_on_response = false;
  bool app_is_running = false;

  Future<String> response_handler() async {
    // catch the response code and update state accordingly
    setState(() {waiting_on_response = true;});
    String response = await stream.first;
    setState(() {waiting_on_response = false;});
    return response;
  }

  Future<String> start_app() async {
    // tell control daemon to start current app
    channel.sink.add('start_app');
    String response = await response_handler();
    await is_running(_selected_app.name);
    return response;
  }

  Future<String> stop_app() async {
    // tell control daemon to stop current app
    channel.sink.add('stop_app');
    String response = await response_handler();
    await is_running(_selected_app.name);
    return response;
  }

  Future<String> get_app() async {
    // ask control daemon for currnet app data
    channel.sink.add('get_app');
    return await response_handler();
  }

  Future<String> load_app(String app_name) async {
    // ask control daemon to load new app
    channel.sink.add('load_app');
    channel.sink.add(app_name);
    String response = await response_handler();
    await update_app_data();
    return response;
  }

  Future<void> is_running(String app_name) async {
    // check if given app is currently running
    if (app_name != "No app selected") {
      channel.sink.add('is_running');
      channel.sink.add(app_name);
      String response = await response_handler();
      setState(() {
        app_is_running = (response == '1');
      });
    }
  }

  Future<void> update_app_data() async {
    // update currently displayed app data
    String response = await get_app();
    if (response.length > 0) {
      var recieved_json = jsonDecode(response);
      setState(() {
        _selected_app = AppMetadata.from_json(recieved_json);
      });
      is_running(_selected_app.name);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: waiting_on_response
          ? Padding(
              padding: EdgeInsets.all(16),
              child: CircularProgressIndicator(),
            )
          : normal_view(context),
      ),
    );
  }

  Widget normal_view(BuildContext context) {
    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: <Widget>[
        Text(
          _selected_app.name,
          style: Theme.of(context).textTheme.headline4,
        ),
        Padding(
          padding: EdgeInsets.all(16),
          child: app_is_running
            ? FloatingActionButton(
                onPressed: stop_app,
                tooltip: 'Stop',
                child: const Icon(Icons.stop),
              )
            : FloatingActionButton(
                onPressed: start_app,
                tooltip: 'Run',
                child: const Icon(Icons.play_arrow_outlined),
              ),
        ),
        Text(
          _selected_app.note
        ),
      ],
    );
  }

  @override
  void initState() {
    super.initState();
    update_app_data();
  }
}
