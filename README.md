# Blackout.go

* leveraging gmer driver to effectively disabling or killing EDRs and AVs.
* it bypass HVCI fluently
* the sample is sourced from loldrivers https://www.loldrivers.io/drivers/7ce8fb06-46eb-4f4f-90d5-5518a6561f15/
# usage

* Place the driver `Blackout.sys` in the same path as the executable
* The executable should be run in the context of an administrator
* Blackout.exe -p <process_id>
* for windows defender keep the program running to prevent the service from restarting it