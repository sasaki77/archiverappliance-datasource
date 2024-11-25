.. archiverappliance-datasource documentation master file, created by
   sphinx-quickstart on Thu Dec  5 14:10:38 2019.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

Archiver Appliance Datasouce
==========================================
Archiver Appliance Datasource is a Grafana plugin to visualize archived PV data in Archiver Appliance.
EPICS Archiver Appliance is an archive engine for EPICS control systems. See `Archiver Appliance site <https://epicsarchiver.readthedocs.io>`_ for more information about Archiver Appliance.

Links
-----
* `Source(GitHub) <https://github.com/sasaki77/archiverappliance-datasource>`_
* `Archiver Appliance site <https://epicsarchiver.readthedocs.io>`_

Features
--------
* Select multiple PVs by using Regex (Only supports wildcard pattern like ``PV.*`` and alternation pattern like ``PV(1|2)``)
* Legend alias with regular expression pattern
* Data retrieval with data processing (See `Archiver Appliance User Guide <https://epicsarchiver.readthedocs.io/en/latest/user/userguide.html#processing-of-data>`_ for processing of data)
* Using PV names for Grafana variables
* Transform your data with processing functions
* Live update with stream feature
* Find and notify problems with alerting feature

Documentation
-------------
.. toctree::
   :maxdepth: 2

   installation
   configuration
   query
   variables
   functions
   tips
   development
   releases
