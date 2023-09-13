import webui.ui.tabs.training.rvc as rvc
import webui.ui.tabs.training.training.rvc_workspace as rvc_ws
import json

# From webui/ui/tabs/training/rvc.py
def load_workspace(name):
    rvc_ws.current_workspace = rvc_ws.RvcWorkspace(name).load()
    ws = rvc_ws.current_workspace

# From webui/ui/tabs/training/rvc.py
def create_workspace(name, vsr):
    rvc_ws.current_workspace = rvc_ws.RvcWorkspace(name).create({
        'vsr': vsr
    })
    rvc_ws.current_workspace.save()
    load_workspace(name)

def main():
	with open('rvc-train-config.json') as f:
		args = json.load(f)

	create_workspace(args['model'], "v2 40k")

	rvc.change_setting("dataset", args['src_dir'])
	gen_object = rvc_ws.process_dataset()
	for i in gen_object:
		print(i)

	rvc.change_setting("f0", args['alg'])
	gen_object = rvc_ws.pitch_extract()
	for i in gen_object:
		print(i)

	gen_object = rvc_ws.create_index()
	for i in gen_object:
		print(i)

	rvc.change_setting("batch_size", args['batch_size'])
	rvc.change_setting("save_epochs", args['epochs'])
	gen_object = rvc_ws.train_model("f0", args['epochs'])
	for i in gen_object:
		print(i)

if __name__ == "__main__":
    main()
