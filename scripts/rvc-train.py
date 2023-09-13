import webui.ui.tabs.training.rvc as rvc
import webui.ui.tabs.training.training.rvc_workspace as rvc_ws

def arg_parse() -> tuple:
    parser = argparse.ArgumentParser()
    parser.add_argument("--model", type=str, help="model name")
    parser.add_argument("--src_dir", type=str, help="training data source directory")
    parser.add_argument("--alg", type=str, default="harvest", help="train algorithm")
    parser.add_argument("--batch_size", type=int, default=12, help="batch size")
    parser.add_argument("--epochs", type=int, default=100, help="epochs")

    args = parser.parse_args()
    sys.argv = sys.argv[:1]

    return args

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
	args = arg_parse()

	create_workspace(args.model, "v2 40k")

	rvc.change_setting("dataset", args.src_dir)
	gen_object = rvc_ws.process_dataset()
	for i in gen_object:
		print(i)

	rvc.change_setting("f0", args.alg)
	gen_object = rvc_ws.pitch_extract()
	for i in gen_object:
		print(i)

	gen_object = rvc_ws.create_index()
	for i in gen_object:
		print(i)

	rvc.change_setting("batch_size", args.batch_size)
	rvc.change_setting("save_epochs", args.epochs)
	gen_object = rvc_ws.train_model("f0", args.epochs)
	for i in gen_object:
		print(i)

if __name__ == "__main__":
    main()
