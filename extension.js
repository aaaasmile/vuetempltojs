// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
const vscode = require('vscode');
const fs = require('fs');
const path = require('path');
const child_process = require('child_process');

// this method is called when your extension is activated
// your extension is activated the very first time the command is executed

/**
 * @param {vscode.ExtensionContext} context
 */
function activate(context) {

	// Use the console to output diagnostic information (console.log) and errors (console.error)
	// This line of code will only be executed once when your extension is activated
	console.log('"vuetempltojs" is now active!');

	// The command has been defined in the package.json file
	// Now provide the implementation of the command with  registerCommand
	// The commandId parameter must match the command field in package.json
	let disposable = vscode.commands.registerCommand('extension.vueTemplInJs', function () {
		// The code you place here will be executed every time your command is executed

		runcmd(context);
	
		// Display a message box to the user
		//vscode.window.showInformationMessage('Here I want to copy the template in component');
	});

	context.subscriptions.push(disposable);
}
exports.activate = activate;

// this method is called when your extension is deactivated
function deactivate() { }

function runcmd(context) {
	let tool = context.asAbsolutePath("TextProc\\TextProc.exe"); // nel package vie incluso anche TextProc.exe nella sottodir TextProc
	//console.log("** Logpath is ", context.logPath)
	
	const uri = vscode.window.activeTextEditor.document.uri;
	let fileIsWrong = true
	let fname = vscode.window.activeTextEditor.document.fileName
	if (uri.scheme === 'file'){
		let extName = path.extname(fname)
		fileIsWrong = (extName !== '.vue')
	}
	
	//console.log('** file', fname)
	if (fileIsWrong){
		console.log('This command is available only for .vue files as active document')
		return
	}

	if (vscode.window.activeTextEditor.document.isDirty){
		console.log('Save the file before processing it')
		vscode.window.activeTextEditor.document.save()
	}
	
	let args = ['-vue', fname, '-logpath', context.logPath]
	let cwd = path.dirname(fname)

	//console.log(tool, args)

	try {
		console.log("Starting the TextProc")
		let textProc = child_process.spawn(tool, args, {cwd} );
		textProc.on('error', (err) => {
			console.error('Failed to start subprocess.', err);
		});

		textProc.stdout.on('data', (data) => {
			console.log(`stdout: ${data}`);
		});
		ls.stderr.on('data', (data) => {
			console.error(`stderr: ${data}`);
		});
		ls.on('close', (code) => {
			console.log(`child process exited with code ${code}`);
		});
		console.log("TextProc executed.")

	} catch (err) {
		vscode.window.showErrorMessage(err);
		return;
	}
}


module.exports = {
	activate,
	deactivate
}
