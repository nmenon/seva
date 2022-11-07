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

class WebSocketCommand {
  String command;
  List<String> arguments;
  int exit_code = 0;
  List<String> response = [];

  WebSocketCommand(this.command, this.arguments, this.exit_code, this.response);

  WebSocketCommand.outbound(this.command, this.arguments);

  WebSocketCommand.from_json(Map<String, dynamic> json)
      : command = json['command'],
        arguments = json['arguments'],
        exit_code = json['exit_code'],
        response = json['response'];

  Map<String, dynamic> to_json() => {
        'command': command,
        'arguments': arguments,
        'exit_code': exit_code,
        'response': response,
      };

  void send() {
    channel.sink.add(jsonEncode(this.to_json()));
  }
}

class WebSocketStatusState extends State<WebSocketStatus> {
  AppMetadata _selected_app = AppMetadata.empty();
  bool waiting_on_response = false;
  bool app_is_running = false;

  Future<WebSocketCommand> response_handler() async {
    // catch the response code and update state accordingly
    setState(() {
      waiting_on_response = true;
    });
    String response = await stream.first;
    setState(() {
      waiting_on_response = false;
    });
    return WebSocketCommand.from_json(jsonDecode(response));
  }

  Future<WebSocketCommand> start_app() async {
    // tell control daemon to start current app
    WebSocketCommand.outbound('start_app', []).send();
    WebSocketCommand command = await response_handler();
    await is_running(_selected_app.name);
    return command;
  }

  Future<WebSocketCommand> stop_app() async {
    // tell control daemon to stop current app
    WebSocketCommand.outbound('stop_app', []).send();
    WebSocketCommand command = await response_handler();
    await is_running(_selected_app.name);
    return command;
  }

  Future<WebSocketCommand> get_app() async {
    // ask control daemon for currnet app data
    WebSocketCommand.outbound('get_app', []).send();
    return await response_handler();
  }

  Future<WebSocketCommand> load_app(String app_name) async {
    // ask control daemon to load new app
    WebSocketCommand.outbound('load_app', [app_name]).send();
    WebSocketCommand command = await response_handler();
    await update_app_data();
    return command;
  }

  Future<void> is_running(String app_name) async {
    // check if given app is currently running
    if (app_name != "No app selected") {
      WebSocketCommand.outbound('is_running', [app_name]).send();
      WebSocketCommand command = await response_handler();
      setState(() {
        app_is_running = (command.response[0] == '1');
      });
    }
  }

  Future<void> update_app_data() async {
    // update currently displayed app data
    WebSocketCommand command = await get_app();
    if (command.response.length > 0) {
      var recieved_json = jsonDecode(command.response[0]);
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
        Card(
          child: Container(
            constraints:
                BoxConstraints(minHeight: 150, maxHeight: 150, maxWidth: 800),
            child: Row(
              children: <Widget>[
                Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: <Widget>[
                    Padding(
                      padding: EdgeInsets.all(16),
                      child: Container(
                        color: Colors.grey,
                        height: 118,
                        width: 118,
                        child: Padding(
                          padding: EdgeInsets.all(16),
                          child: Icon(Icons.auto_awesome),
                        ),
                      ),
                    ),
                  ],
                ),
                Expanded(
                  flex: 2,
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.start,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Padding(
                        padding: EdgeInsets.all(16),
                        child: Text(
                          _selected_app.name,
                          style: Theme.of(context).textTheme.headline4,
                        ),
                      ),
                      Padding(
                        padding:
                            EdgeInsets.only(left: 16, right: 16, bottom: 16),
                        child: Text(_selected_app.note),
                      ),
                    ],
                  ),
                ),
                Column(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: <Widget>[
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
                  ],
                ),
              ],
            ),
          ),
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
