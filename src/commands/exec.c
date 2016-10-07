/*
 * This file is part of cc-oci-runtime.
 *
 * Copyright (C) 2016 Intel Corporation
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation; either version 2
 * of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
 */

#include "command.h"
#include "oci-config.h"

extern struct start_data start_data;

/* ignore -pedantic to cast handle_option_console, a function pointer, to a
 * void* */
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wpedantic"
static GOptionEntry options_exec[] =
{
	{
		"console", 0, G_OPTION_FLAG_OPTIONAL_ARG,
		G_OPTION_ARG_CALLBACK, handle_option_console,
		"set pty console that will be used in the container",
		NULL
	},
	{
		"process", 'p' , G_OPTION_FLAG_NONE,
		G_OPTION_ARG_STRING, &start_data.process_json,
		"specify path to process.json",
		NULL
	},
	{
		"detach", 'd' , G_OPTION_FLAG_NONE,
		G_OPTION_ARG_NONE, &start_data.detach,
		"exec process in detach mode",
		NULL
	},
	{
		"pid-file", 0, G_OPTION_FLAG_NONE,
		G_OPTION_ARG_STRING, &start_data.pid_file,
		"the file to write the process ID of the new "
		"process executed in the container",
		NULL
	},
	{NULL}
};

static gboolean
handler_exec (const struct subcommand *sub,
		struct cc_oci_config *config,
		int argc, char *argv[])
{
	struct cc_oci_process_exec  exec_process = { {0} };
	struct oci_state            *state = NULL;
	gchar                       *config_file = NULL;
	gboolean                    ret;

	g_assert (sub);
	g_assert (config);

	if (handle_default_usage (argc, argv, sub->name,
				&ret, 1, "<cmd> [args]")) {
		goto out;
	}

	/* Used to allow us to find the state file */
	config->optarg_container_id = argv[0];

	/* Jump over the container name */
	argv++; argc--;

	/* Verify that exits process information (args or process.json)*/
	if ( argc < 1  && ! start_data.process_json) {
		g_print ("Usage: %s <container-id> <cmd> [args]\n",
			sub->name);
		goto out;

	}

	ret = cc_oci_get_config_and_state (&config_file, config, &state);
	if (! ret) {
		goto out;
	}

	/* Copy args and options to exec_process */
	if(argc > 0) {
		exec_process.process.args = g_new0 (gchar *, (gsize)argc + 1);
	}

	for (int  i = 0; i < argc; ++i) {
	       exec_process.process.args[i] = g_strdup (argv[i]);
	       if (! exec_process.process.args[i]){
		       goto out;
	       }
	}

	exec_process.detach = start_data.detach;
	exec_process.pid_file = g_strdup(start_data.pid_file);
	exec_process.console = g_strdup(start_data.console);
	exec_process.process_json_file = g_strdup(start_data.process_json);

	ret = cc_oci_exec (config, state, &exec_process);
	if (! ret) {
		goto out;
	}

	ret = true;

out:
	g_free_if_set (config_file);
	cc_oci_state_free (state);
	cc_oci_process_exec_free(&exec_process);

	return ret;
}

struct subcommand command_exec =
{
	.name        = "exec",
	.options     = options_exec,
	.handler     = handler_exec,
	.description = "execute a new task inside an existing container",
};
