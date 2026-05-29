package hca

import "eman1can/enum"

func ChannelInit(channels uint, channelConfig uint, trackCount uint, baseBandCount uint, stereoBandCount uint) []StChannel {
	channel := make([]StChannel, channels)

	channelsPerTrack := channels / trackCount
	if stereoBandCount > 0 && channelsPerTrack > 1 {
		ct := 0
		for ix := 0; ix < int(trackCount); ix++ {
			switch channelsPerTrack {
			case 2:
				channel[ct+0].Type = enum.StereoPrimary
				channel[ct+1].Type = enum.StereoSecondary
			case 3:
				channel[ct+0].Type = enum.StereoPrimary
				channel[ct+1].Type = enum.StereoSecondary
				channel[ct+2].Type = enum.Discrete
			case 4:
				channel[ct+0].Type = enum.StereoPrimary
				channel[ct+1].Type = enum.StereoSecondary
				if channelConfig == 0 {
					channel[ct+2].Type = enum.StereoPrimary
					channel[ct+3].Type = enum.StereoSecondary
				} else {
					channel[ct+2].Type = enum.Discrete
					channel[ct+3].Type = enum.Discrete
				}
			case 5:
				channel[ct+0].Type = enum.StereoPrimary
				channel[ct+1].Type = enum.StereoSecondary
				channel[ct+2].Type = enum.Discrete
				if channelConfig == 0 {
					channel[ct+3].Type = enum.StereoPrimary
					channel[ct+4].Type = enum.StereoSecondary
				} else {
					channel[ct+3].Type = enum.Discrete
					channel[ct+4].Type = enum.Discrete
				}
			case 6:
				channel[ct+0].Type = enum.StereoPrimary
				channel[ct+1].Type = enum.StereoSecondary
				channel[ct+2].Type = enum.Discrete
				channel[ct+3].Type = enum.Discrete
				channel[ct+4].Type = enum.StereoPrimary
				channel[ct+5].Type = enum.StereoSecondary
			case 7:
				channel[ct+0].Type = enum.StereoPrimary
				channel[ct+1].Type = enum.StereoSecondary
				channel[ct+2].Type = enum.Discrete
				channel[ct+3].Type = enum.Discrete
				channel[ct+4].Type = enum.StereoPrimary
				channel[ct+5].Type = enum.StereoSecondary
				channel[ct+6].Type = enum.Discrete
			case 8:
				channel[ct+0].Type = enum.StereoPrimary
				channel[ct+1].Type = enum.StereoSecondary
				channel[ct+2].Type = enum.Discrete
				channel[ct+3].Type = enum.Discrete
				channel[ct+4].Type = enum.StereoPrimary
				channel[ct+5].Type = enum.StereoSecondary
				channel[ct+6].Type = enum.StereoPrimary
				channel[ct+7].Type = enum.StereoSecondary
			default:
				// Implied all Discrete (0)
			}
			ct += int(channelsPerTrack)
		}
	}

	for ix := 0; ix < int(channels); ix++ {
		if channel[ix].Type != enum.StereoSecondary {
			channel[ix].CodedCount = baseBandCount + stereoBandCount
		} else {
			channel[ix].CodedCount = baseBandCount
		}
	}

	return channel
}
