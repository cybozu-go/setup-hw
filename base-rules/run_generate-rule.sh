#!/usr/bin/bash
#
#  Shell for collector generate-rule execute
#
#  Usage: run_generate-rule.sh INPUT_FILES... > OUTPUT_FILE
#

SCRIPT_DIR="$(cd $(dirname $0); pwd)"

collector generate-rule \
  --base-rule=${SCRIPT_DIR}/dell.yaml \
  --key=AddressParityError:bool \
  --key=CorrectableECCError:bool \
  --key=DataLossDetected:bool \
  --key=FailurePredicted:bool \
  --key=Health:health \
  --key=PowerConsumedWatts:number \
  --key=PredictedMediaLifeLeftPercent:number \
  --key=Reading:number \
  --key=ReadingCelsius:number \
  --key=SpareBlock:bool \
  --key=State:state \
  --key=Temperature:bool \
  --key=UncorrectableECCError:bool \
  $@
