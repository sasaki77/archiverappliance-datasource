.. archiverappliance-datasource documentation master file, created by
   sphinx-quickstart on Thu Dec  5 14:10:38 2019.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

Archiver Appliance Datasouce
==========================================
Archiver Appliance Datasource is a Grafana plugin to visualize archived PV data in Archiver Appliance.
EPICS Archiver Appliance is an archive engine for EPICS control systems. See `Archiver Appliance site <https://domain.invalid/>`_ for more information about Archiver Appliance.

Features
--------
* Select multiple PVs by using Regex (Only supports wildcard pattern like ``PV.*`` and alternation pattern like ``PV(1|2)``)
* Legend alias with regex pattern
* Data retrieval with data processing (See `Archiver Appliance User Guide <https://slacmshankar.github.io/epicsarchiver_docs/userguide.htm>`_ for processing of data)
* Using PV names for Grafana variables
* Transform your data with processing functions

Documentation
-------------
.. toctree::
   :maxdepth: 2

   installation
   configuration
   query
   variables
   functions

