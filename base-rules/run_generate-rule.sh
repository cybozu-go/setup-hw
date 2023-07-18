#!/usr/bin/bash
#
#  Shell for collector generate-rule execute on container
#
#   2023/7/14  RedFish v1.17
#

# Please edit for your environment
RULE="dell-redfish-v117.yaml"         # Base rule
INPUT1="r7525-data-omsa1100.json"     # Output by "collector show" command 
INPUT2="r6525-data-omsa1100.json"     #  for each PowerEdge Model
OUTPUT="dell_redfish_1.17.0.yml"      # Output rule file for Promethus/Grafana 
SETUP_HW="quay.io/neco_test/setup-hw:dev"  # Latest container image if there is dell command upgrade. 

docker run -it --name=setup-hw \
   -v ${PWD}:/mnt \
   --rm \
   ${SETUP_HW} collector generate-rule \
     --base-rule=/mnt/${RULE} \
     --key=Health:health \
     --key=State:state \
     --key=FailurePredicted:bool \
     --key=PredictedMediaLifeLeftPercent:number \
     --key=AddressParityError:bool \
     --key=CorrectableECCError:bool \
     --key=SpareBlock:bool \
     --key=Temperature:bool \
     --key=UncorrectableECCError:bool \
     --key=DataLossDetected:bool \
     --key=ReadingCelsius:number \
     --key=PowerConsumedWatts:number \
     --key=Reading:number \
   /mnt/${INPUT1} /mnt/${INPUT2} > ${OUTPUT}
